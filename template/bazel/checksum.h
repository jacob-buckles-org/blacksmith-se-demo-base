#ifndef ANVIL_TOOLS_CHECKSUM_H_
#define ANVIL_TOOLS_CHECKSUM_H_

#include <cstdint>
#include <string>

namespace anvil {

// Fletcher-64 checksum used to verify export archive integrity.
uint64_t Fletcher64(const std::string& data);

// Formats a checksum the way export manifests expect: 16 lowercase hex chars.
std::string FormatChecksum(uint64_t checksum);

}  // namespace anvil

#endif  // ANVIL_TOOLS_CHECKSUM_H_
