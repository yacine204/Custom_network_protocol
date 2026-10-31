#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <stdint.h>
#include <arpa/inet.h>
#include <stdbool.h>
#include <sys/socket.h>


typedef struct __attribute__((__packed__)) {
    uint8_t  version_ihl;     
    uint8_t  tos;              
    uint16_t total_length;    
    uint16_t id;               
    uint16_t flags_fragment;   
    uint8_t  ttl;              
    uint8_t  protocol;         
    uint16_t header_checksum;  
    uint32_t src_ip;           
    uint32_t dest_ip;          
} IPv4Header;

uint16_t compute_checksum(uint16_t *addr, int count) {
    uint32_t sum = 0;
    while (count > 1) {
        sum += *addr++;
        count -= 2;
    }
    if (count > 0) {
        sum += *(uint8_t *)addr;
    }
    while (sum >> 16) {
        sum = (sum & 0xFFFF) + (sum >> 16);
    }

    return (uint16_t)~sum;
}

char* Encapsulate(char *payload, char *dstip, int *out_size) {
    size_t payload_len = strlen(payload);
    if (payload_len == 0){
        printf("[ERROR] payload len == 0.\n");
        return NULL;
    }

    int packet_size = sizeof(IPv4Header) + payload_len;
    char *packet = malloc(packet_size); 
    
    if (packet == NULL){
        printf("[ERROR] couldnt alloc for capsule\n");
        return NULL; 
    }

    IPv4Header* capsule = (IPv4Header*) packet;

    capsule->version_ihl = 0x45;
    capsule->tos = 0;
    capsule->total_length = htons(packet_size);
    capsule->id = htons(0);
    capsule->flags_fragment = htons(0);
    capsule->ttl = 64;
    capsule->protocol = 253;
    capsule->header_checksum = 0;
    capsule->src_ip = inet_addr("10.0.0.10");
    capsule->dest_ip = inet_addr(dstip);
    capsule->header_checksum = compute_checksum((uint16_t*)capsule, sizeof(IPv4Header));
    
    memcpy(packet + sizeof(IPv4Header), payload, payload_len);
    
    *out_size = packet_size; 
    return packet;
}


int main(int argc, char* argv[]){
    if(argc < 2){
        printf("Usage: clt <ip>\n");
        return 1;
    }
    
    char *dstIp = malloc(strlen(argv[1]) + 1); 
    if (dstIp == NULL) return 1;
    strcpy(dstIp, argv[1]);
    
    printf("[INFO] Trying to connect to %s\n", dstIp);

    char buffer[1024];

    int fd = socket(AF_INET, SOCK_RAW, 253);

    if(fd<0){
        perror("[ERROR] Socket creation failed\n");
        free(dstIp);
        return 1;
    }


    int one = 1;

    if (setsockopt(fd, IPPROTO_IP, IP_HDRINCL, &one, sizeof(one)) < 0){
        perror("[ERROR] Error setting IP_HDRINCL\n");
        close(fd);
        free(dstIp);
        return 1;
    }

    struct sockaddr_in dest_addr;
    memset(&dest_addr, 0, sizeof(dest_addr));
    dest_addr.sin_family = AF_INET;

    if (inet_pton(AF_INET, dstIp, &dest_addr.sin_addr)<=0){
        printf("[ERR] Invalid ip\n");
        close(fd);
        free(dstIp);
        return 1;
    }

    if (connect(fd, (struct sockaddr*)&dest_addr, sizeof(dest_addr))<0){
        perror("Error connecting\n");
        close(fd);
        return 1;
    }else{
        printf("[INFO] Connected to %u\n",dest_addr.sin_addr.s_addr);
    }


    while(true){
        if(fgets(buffer, sizeof(buffer), stdin) != NULL) {
            buffer[strcspn(buffer, "\n")] = '\0'; 
        } else {
            break; 
        }

        int packet_size = 0;
        char* packet = Encapsulate(buffer, dstIp, &packet_size);

        if(packet==NULL){
            continue;
        }

        ssize_t sent = send(fd, packet, packet_size, 0);
        if (sent < 0) {
            perror("[ERROR] send failed");
        } else if ((size_t)sent != (size_t)packet_size) {
            fprintf(stderr, "[WARN] partial send: %zd/%d\n", sent, packet_size);
        } else {
            printf("[INFO] Sent %zd bytes\n", sent);
        }
        if (packet != NULL) {
           
            printf("[DEBUG] Packet Sent (%d bytes): ", packet_size);
            for(int i = 0; i < packet_size; i++) {
                printf("%02X ", (unsigned char)packet[i]);
            }
            printf("\n");
            
            free(packet); 
        }
    }

    free(dstIp);
    close(fd);
    return 0;
}
