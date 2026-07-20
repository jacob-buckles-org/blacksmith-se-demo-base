#include <fstream>
#include <iostream>
#include <sstream>

#include "checksum.h"

int main(int argc, char** argv) {
  if (argc != 2) {
    std::cerr << "usage: checksum_tool <file>" << std::endl;
    return 2;
  }
  std::ifstream in(argv[1], std::ios::binary);
  if (!in) {
    std::cerr << "cannot open " << argv[1] << std::endl;
    return 1;
  }
  std::ostringstream contents;
  contents << in.rdbuf();
  std::cout << metrics::FormatChecksum(metrics::Fletcher64(contents.str())) << std::endl;
  return 0;
}
