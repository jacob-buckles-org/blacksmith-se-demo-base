#include "checksum.h"

#include <cstdio>

namespace metrics {

uint64_t Fletcher64(const std::string& data) {
  uint64_t sum1 = 0;
  uint64_t sum2 = 0;
  for (unsigned char c : data) {
    sum1 = (sum1 + c) % 4294967295ULL;
    sum2 = (sum2 + sum1) % 4294967295ULL;
  }
  return (sum2 << 32) | sum1;
}

std::string FormatChecksum(uint64_t checksum) {
  char buf[17];
  std::snprintf(buf, sizeof(buf), "%016llx", static_cast<unsigned long long>(checksum));
  return std::string(buf);
}

}  // namespace metrics
