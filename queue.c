#include <stdio.h>
#include <poll.h>
#include "./honeyshell.h"

int wait_n_read(int fd, char *buf) {
    struct pollfd pfd = {
        .fd      = fd,
        .events  = POLLIN,
        .revents = 0
    };

    poll(&pfd, 1, -1);

    int ret = read(fd, buf, sizeof(char[512]) - 1);

    if (ret == -1) {
        perror("In read()");
    }

    return ret;
}

char *wait_for_creds(auth_queue *queue) {
    char *val = malloc(sizeof(char) * 512);

    if (wait_n_read(queue->chan[0], val) == -1) {
        return NULL;
    }

    return val;
}

auth_queue create_auth_queue() {
    auth_queue queue = {
        .chan = {0, 0}
    };

    return queue;
}
