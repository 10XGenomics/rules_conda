// Demonstrate linking against a conda package, in this case OpenSSL.

#include <openssl/sha.h>

#include <iostream>
#include <string_view>

using std::string_view;

int main() {
  const string_view message = "Hello World";
  unsigned char hash[SHA256_DIGEST_LENGTH];

  SHA256_CTX sha256;
  SHA256_Init(&sha256);
  SHA256_Update(&sha256, message.data(), message.size());
  SHA256_Final(hash, &sha256);

  std::cout << "SHA-256 Hash of 'Hello World': ";
  for (int i = 0; i < SHA256_DIGEST_LENGTH; i++) {
    printf("%02x", hash[i]);
  }
  std::cout << std::endl;

  return 0;
}
