// Demonstrate linking against a conda package, in this case OpenSSL.

#include <openssl/evp.h>

#include <iostream>
#include <string_view>

using std::string_view;

int main() {
  const string_view message = "Hello World";
  unsigned char hash[EVP_MAX_MD_SIZE];
  size_t hashLen;

  if (!EVP_Q_digest(nullptr, "SHA256", nullptr, message.data(),
                    message.length(), hash, &hashLen)) {
    return 1;
  }

  std::cout << "SHA-256 Hash of 'Hello World': ";
  for (int i = 0; i < hashLen; i++) {
    printf("%02x", hash[i]);
  }
  std::cout << std::endl;

  return 0;
}
