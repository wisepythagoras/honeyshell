#include <stdio.h>
#include <poll.h>
#include "./honeyshell.h"

int wait_n_read(int fd, char *buf, int szBuf) {
    struct pollfd pfd = {
        .fd      = fd,
        .events  = POLLIN,
        .revents = 0
    };

    poll(&pfd, 1, -1);

    int ret = read(fd, buf, szBuf);

    if (ret == -1) {
        perror("In read()");
    }

    return ret;
}

char *wait_for_password(struct session_data_struct *sdata) {
    char *val = malloc(sizeof(char) * 256);
printf("%i - %i\n", sdata->chan[READ_FD], sdata->chan[1]);

    if (wait_n_read(sdata->chan[READ_FD], val, sizeof(val) - 1) == -1) {
        printf("Failed to read\n");
        return val;
    }

    return val;
}
