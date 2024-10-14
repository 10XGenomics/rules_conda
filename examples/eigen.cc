// Trivial test case to show getting headers from a conda package.

#include <Eigen/Dense>
#include <iostream>

using Eigen::Matrix;
using std::cout;
using std::endl;

int main() {
  Matrix<float, 3, 3> matrixA;
  matrixA.setZero();
  cout << matrixA << endl;
  return 0;
}