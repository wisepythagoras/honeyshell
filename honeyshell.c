#include <stdio.h>
#include <poll.h>
#include <ctype.h>
#include "./honeyshell.h"

char *escape(char *str) {
    char *escaped = malloc(sizeof(char) * strlen(str) * 2);
    memset(escaped, 0, strlen(str) * 2);

    for (int i = 0; i < strlen(str); i++) {
        if (str[i] == '\\') {
            strcat(escaped, "\\\\");
        } else if (str[i] == '"') {
            strcat(escaped, "\\\"");
        } else {
            sprintf(escaped, "%s%c", escaped, str[i]);
        }
    }

    return escaped;
}

const char *get_ssh_key_type(const ssh_key key) {
	if (key == NULL) {
		return NULL;
	}

	return key->type_c;
}

static ssh_channel channel_open(ssh_session session, void *userdata) {
    struct session_data_struct *sdata = (struct session_data_struct *) userdata;

    sdata->channel = ssh_channel_new(session);
    return sdata->channel;
}

static int auth_publickey(ssh_session session,
                          const char *user,
                          struct ssh_key_struct *pubkey,
                          char signature_state,
                          void *userdata)
{
    struct session_data_struct *sdata = (struct session_data_struct *) userdata;

    (void) user;
    (void) session;

    // States include:
    // - SSH_PUBLICKEY_STATE_NONE
    // - SSH_PUBLICKEY_STATE_VALID
    //
    // Response types:
    // - SSH_AUTH_SUCCESS
    // - SSH_AUTH_DENIED

    // We don't actually want to accept any key here, we only want to log them.
    sdata->authenticated = 0;

    return SSH_AUTH_DENIED;
}

static int auth_password(ssh_session session, const char *user,
                         const char *pass, void *userdata) {
    struct session_data_struct *sdata = (struct session_data_struct *) userdata;

    (void) session;

    sdata->auth_attempts++;

    char poll_msg[512];
    sprintf(poll_msg, "[\"%s\",\"%s\",\"%i\"]", escape((char *) user), escape((char *) pass), sdata->auth_attempts);
    write(sdata->queue->chan[1], poll_msg, sizeof(char[256]));

    // Use logic like this to trick bots into thinking they've authenticated.
    // if (strcmp(user, username) == 0 && strcmp(pass, password) == 0) {
    //     sdata->authenticated = 1;
    //     return SSH_AUTH_SUCCESS;
    // }

    return SSH_AUTH_DENIED;
}

void handle_auth(ssh_session session, auth_queue *pqueue) {
    struct session_data_struct sdata = {
        .channel = NULL,
        .auth_attempts = 0,
        .authenticated = 0,
        .queue = pqueue
    };

    pipe(&pqueue->chan[0]);

    struct ssh_server_callbacks_struct server_cb = {
        .userdata = &sdata,
        .auth_password_function = auth_password,
        .channel_open_request_session_function = channel_open,
        .auth_pubkey_function = auth_publickey
    };

    ssh_callbacks_init(&server_cb);
    ssh_set_server_callbacks(session, &server_cb);

    if (ssh_handle_key_exchange(session) != SSH_OK) {
        fprintf(stderr, "%s\n", ssh_get_error(session));
        return;
    }

    ssh_event event = ssh_event_new();

    ssh_event_add_session(event, session);

    int n = 0;

    while (sdata.authenticated == 0 || sdata.channel == NULL) {
        if (sdata.auth_attempts >= 3 || n >= 10 * 100) {
            return;
        }

        int res = ssh_event_dopoll(event, 100);

        if (res == SSH_ERROR) {
            fprintf(stderr, "%s\n", ssh_get_error(session));
            return;
        }

        n++;
    }

    ssh_channel_send_eof(sdata.channel);
    ssh_channel_close(sdata.channel);

    for (n = 0; n < 50 && (ssh_get_status(session) & (SSH_CLOSED | SSH_CLOSED_ERROR)) == 0; n++) {
        ssh_event_dopoll(event, 100);
    }

    pqueue = NULL;
}
