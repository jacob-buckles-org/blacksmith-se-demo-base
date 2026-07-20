#include "checksum.h"

#include <cassert>
#include <iostream>

int main() {
  // Empty input.
  assert(metrics::Fletcher64("") == 0);

  // Deterministic for identical input.
  assert(metrics::Fletcher64("metrics") == metrics::Fletcher64("metrics"));

  // Order-sensitive: transposed bytes must differ.
  assert(metrics::Fletcher64("ab") != metrics::Fletcher64("ba"));

  // Formatting is fixed-width lowercase hex.
  assert(metrics::FormatChecksum(0).size() == 16);
  assert(metrics::FormatChecksum(255) == "00000000000000ff");

  std::cout << "checksum_test: all assertions passed" << std::endl;
  return 0;
}
