// To compile:
// g++ -c hello.cc
// ar r libhello.a hello.o
// sudo install libhello.a /usr/local/lib

#include <iostream>
using namespace std;

class Hello {
 public:
  Hello() {
    cout << "In Hello constructor." << endl;
  }

  void Print() {
    cout << "Hello world!" << endl;
  }
};

extern "C" void hello_world() {
	Hello hello;
	hello.Print();
}
