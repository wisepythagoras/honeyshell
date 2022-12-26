#ifndef HONEYSHELL_H
#define HONEYSHELL_H

#include <stdlib.h>
#include <pwd.h>
#include <sys/types.h>
#include <grp.h>
#include <inttypes.h>
#include <sys/types.h>
#include <libssh/libssh.h>
#include <libssh/server.h>
#include <libssh/callbacks.h>

struct ssh_key_struct {
    enum ssh_keytypes_e type;
    int flags;
    const char *type_c;
	int ecdsa_nid;
#if defined(HAVE_LIBGCRYPT)
    gcry_sexp_t dsa;
    gcry_sexp_t rsa;
    gcry_sexp_t ecdsa;
#elif defined(HAVE_LIBMBEDCRYPTO)
    mbedtls_pk_context *rsa;
    mbedtls_ecdsa_context *ecdsa;
    void *dsa;
#elif defined(HAVE_LIBCRYPTO)
    DSA *dsa;
    RSA *rsa;
# if defined(HAVE_OPENSSL_ECC)
    EC_KEY *ecdsa;
# else
    void *ecdsa;
# endif
#endif
    void *cert;
    enum ssh_keytypes_e cert_type;
};

struct password_auth_attempt_msg_struct {
    ssh_session session;
    const char *user;
    const char *pass;
};
typedef struct password_auth_attempt_msg_struct password_auth_attempt_msg;

typedef struct queue_struct {
    int chan[2];
} auth_queue;

struct session_data_struct {
    ssh_channel channel;
    int auth_attempts;
    int authenticated;
    auth_queue *queue;
};

const char *get_ssh_key_type(const ssh_key key);

void handle_auth(ssh_session session, auth_queue *pqueue);

auth_queue create_auth_queue();
char *wait_for_creds(auth_queue *queue);
void push_password_msg(auth_queue *queue, password_auth_attempt_msg *msg);
password_auth_attempt_msg *get_password_msg(auth_queue *queue);
int is_password_queue_empty(auth_queue *queue);

#endif
