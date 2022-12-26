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

#define READ_FD 0
#define WRITE_FD 0

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

struct session_data_struct {
    ssh_channel channel;
    int auth_attempts;
    int authenticated;
    int chan[2];
};

const char *get_ssh_key_type(const ssh_key key);

struct session_data_struct create_session_data_struct();
void handle_auth(ssh_session session, struct session_data_struct *sdata);
char *wait_for_password(struct session_data_struct *sdata);

#endif
