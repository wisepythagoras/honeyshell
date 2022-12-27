#include <stdio.h>
#include <poll.h>
#include "./honeyshell.h"

void poll_queue(int fd) {
    struct pollfd pfd = {
        .fd      = fd,
        .events  = POLLIN,
        .revents = 0
    };

    poll(&pfd, 1, -1);
}

int wait_read_size(int fd, int *buf) {
    poll_queue(fd);

    int ret = read(fd, buf, sizeof(int));

    if (ret == -1) {
        perror("In read()");
    }

    return 1;
}

int wait_read_json(int fd, char *buf, int size) {
    poll_queue(fd);

    int ret = read(fd, buf, (sizeof(char) * size) + 1);

    if (ret == -1) {
        perror("In read()");
    }

    return 1;
}

char *wait_for_creds(auth_queue *queue) {
    int size = 0;

    wait_read_size(queue->chan[0], &size);

    char *val = malloc(sizeof(char) * size);

    if (wait_read_json(queue->chan[0], val, size) == -1) {
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
