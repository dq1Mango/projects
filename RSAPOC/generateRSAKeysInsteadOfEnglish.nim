import std/random
import std/math
import bigints
import json
import std/rdstdin
from strutils import parseInt

randomize()

proc printHelp() = 
  echo """

Silly little RSA demonstration written in Nim by Thomas Ranney

  ./generateRSAKeysInsteadOfEnglish [options] 

Options:
  -b: int         maximum length of key (in bits) (
  -o: file        name of file (without extension)

Note: Both of these fields will be prompted for if not provided
"""
  quit()

proc badUsage() =
  echo "ERROR: Bad Usage\nsee ./generateRSAKeysInsteadOfEnglish --help for proper usage"
  quit()

proc isPrime(n: int): bool =
  if n <= 1:
    return false
  for i in 2 .. (n + 1) div 2:
    if n %% i == 0:
      return false
  return true

#i think this is slow but idgaf
proc iHateEuclid(a: int, b: int): int =
  var x: int = 1
  var i: int = a
  while i mod b != 1:
    x += 1
    i += a
  
  return x

var bits: int

try:
  bits = parseInt(readLineFromStdin(("Enter key size (in bits): ")))
except ValueError:
  echo "ERROR: Invalid input, keys size must be in bits"

let maxSize = 2 ^ bits
let minSize = 2 ^ (bits - 1)

var tmp: int = 0

while not isPrime(tmp):
  tmp = rand(maxSize /% 2) #for more secure (and realistic) encryption larger values should be favored

let p = tmp

while (not isPrime(tmp)) or tmp == p:
  tmp = rand(minSize /% p .. maxSize /% p) #once again bigger is always better

let q = tmp

let n = p * q

let totient = lcm(p - 1, q - 1)

tmp = 0
while gcd(tmp, totient) != 1:
  tmp = rand(2 .. totient - 1)

let e = tmp

let d = iHateEuclid(e, totient)

proc exponent(a: BigInt, n:int): BigInt =
  result = 1.initBigInt
  for i in 1 .. n:
    result *= a

var test: BigInt = 5.initBigInt
echo test mod n.initBigInt
#assert test ^ e ^ d %% n == test %% n

#let cypher = exponent(test, e) mod n.initBigInt
#echo exponent(cypher, d) mod n.initBigInt

let output = readLineFromStdin("Name of file to write to: ")

let privateKey = %* {
  "n": n,
  "e": e,
  "d": d,
  "p": p,
  "q": q,
}

let publicKey = %* {
  "n": n,
  "e": e,
}

writeFile(output & ".key", $privateKey)
writeFile(output & ".pub", $publicKey)

