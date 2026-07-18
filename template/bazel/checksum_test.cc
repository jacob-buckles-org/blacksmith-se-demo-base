#include "checksum.h"

#include <cassert>
#include <iostream>

int main() {
  // Empty input.
  assert(anvil::Fletcher64("") == 0);

  // Deterministic for identical input.
  assert(anvil::Fletcher64("anvil") == anvil::Fletcher64("anvil"));

  // Order-sensitive: transposed bytes must differ.
  assert(anvil::Fletcher64("ab") != anvil::Fletcher64("ba"));

  // Formatting is fixed-width lowercase hex.
  assert(anvil::FormatChecksum(0).size() == 16);
  assert(anvil::FormatChecksum(255) == "00000000000000ff");

  std::cout << "checksum_test: all assertions passed" << std::endl;
  return 0;
}
