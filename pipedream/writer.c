#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>
#include <linux/fcntl.h>        /* For F_GETPIPE_SZ */

int bytes = 8192;
int interval = 1;

int main(int argc, char** argv) {
    if (argc > 1) {
        if (sscanf(argv[1], "%d", &bytes) != 1) {
            fprintf(stderr, "Bad bytes arg\n");
            exit(1);
        }
    }
    if (argc > 2) {
        if (sscanf(argv[2], "%d", &interval) != 1) {
            fprintf(stderr, "Bad interval arg\n");
            exit(1);
        }
    }
    int pipefd[2];
    if (pipe(pipefd) != 0) {
        perror("pipe");
        exit(1);
    }
    int sz = fcntl(pipefd[1], F_GETPIPE_SZ);
    if (sz == -1) {
        perror("fcntl");
        exit(1);
    }
    printf("System reports pipe size: %d\n", sz);
    pid_t p = fork();
    switch (p) {
      case -1:
        perror("fork");
        exit(1);
      case 0: {
          /* child */
          char fds[128];
          sprintf(fds, "%d", pipefd[0]);
          close(pipefd[1]);
          execl("nonreader", "nonreader", fds, (char*)0);
          perror("execl");
          exit(1);
      }
      default: {
          /* parent */
          close(pipefd[0]);
          char v = 37;
          for (int i=0 ; i< bytes; i++ ) {
              if (i % interval == 0) {
                  printf("%d\n", i);
              }
              if (write(pipefd[1], &v, 1) < 0) {
                  printf("WRITE FAILED\n");
                  break;
              }
          }
          printf("LOOP EXITED\n");
      }
    }
}
