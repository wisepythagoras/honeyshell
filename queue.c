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

    int ret = read(fd, buf, sizeof(char[256]) - 1);

    if (ret == -1) {
        perror("In read()");
    }

    return ret;
}

password_auth_attempt_msg *wait_for_password(password_queue *queue) {
    char *val = malloc(sizeof(char) * 256);

    if (wait_n_read(queue->chan[0], val) == -1) {
        return NULL;
    }

    printf("-> %s\n", val);

    return NULL;
}

password_queue create_password_queue() {
    password_queue queue = {
        .first = NULL,
        .last = NULL,
        .count = 0,
        .chan = {0, 0}
    };

    return queue;
}

int is_password_queue_empty(password_queue *queue) {
    return queue->count == 0; // queue->first == NULL && queue->last == NULL;
}

void push_password_msg(password_queue *queue, password_auth_attempt_msg *msg) {
    if (queue == NULL) {
        return;
    }

    password_queue_node new_node = {
        .msg = msg
    };

    queue->count++;

    if (queue->last == NULL || queue->count == 1) {
        queue->last = &new_node;
        queue->first = &new_node;
        return;
    }

    queue->last->next_node = &new_node;
    queue->last = &new_node;
}

password_auth_attempt_msg *get_password_msg(password_queue *queue) {
    if (queue->first == NULL || queue->count == 0) {
        return NULL;
    }

    queue->count--;

    password_auth_attempt_msg *msg = queue->first->msg;
    queue->first = queue->first->next_node;

    if (queue->first == NULL || queue->count == 0) {
        queue->last = NULL;
    }

    return msg;
}
