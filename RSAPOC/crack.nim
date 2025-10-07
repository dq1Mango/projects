import json
import os
import std/rdstdin
import math
import std/times
from std/strutils import parseInt, normalize

proc printHelp() =
  echo """

Silly little RSA cracking demonstration written in Nim by Thomas Ranney

  ./crack [publicKey] [options]

Options:
  -o: file        file to write private key to (defaults to stdout)

"""
  quit()

proc badUsage() =
  echo "ERROR: Bad Usage\nrun ./crack --help for help"
  quit()

let args = commandLineParams()
let length = args.len

type 
  Public = object
    n: int
    e: int

var key: Public
var output: string = ""

if length == 0:
  badUsage()

elif length < 2:
  if args[0] == "--help":
    printHelp()
  else:
    try:
      key = to(parseJson(readFile(args[0])), Public)
    except IOError:
      echo "ERROR: Unable to open file \"" & args[0] & "\""
      quit()
    except ValueError:
      echo "ERROR: Provided public key in incorrect format, run ./generateRSAKeysInsteadOfEnglish to generate a new key"
      quit()

if length == 3:
  if args[1] != "-o":
    badUsage()
  output = args[2]

if length > 3:
  badUsage()


proc primeSieve(n: int): seq[int] =
  echo "Sieving for primes ..."
  var primes: seq[bool] = @[]
  for i in 0 .. n:
    primes.add(true)
  primes[0] = false
  primes[1] = false

  for i in 2..n /% 2:
    for j in 2 * i..n:
      if j mod i == 0:
        primes[j] = false

  var realPrimes: seq[int] = @[]
  for i in 2 .. n:
    if primes[i]:
      realPrimes.add(i)

  echo "Found em ✔️\n"
  return realPrimes

proc factorPrimes(n: int): tuple =
  let primes = primeSieve(n)
  var p: int = 0
  var q: int = primes.len - 1
  
  echo "Factoring..."
  var product: int = primes[p] * primes[q]
  while product != n: #mom look at me i did a binary search
    if product < n:
      p += 1
    else:
      q -= 1
    
    product = primes[p] * primes[q]
  
  echo "Keyser would be proud ✔️\n"
  return (primes[p], primes[q])

proc modularMultapilicativeInverse(a: int, b: int): int =
  var x: int = 1
  var i: int = a
  while i mod b != 1:
    x += 1
    i += a
  
  return x

let n = key.n
let e = key.e

echo "Starting decrytption...\n"
let start = cpuTime()

let tmp: tuple = factorPrimes(n)
let p = tmp[0]
let q = tmp[1]

echo "Computing Totient..."
let totient = lcm(p - 1, q - 1)
echo "That was easy ✔️\n"

echo "Doing math un-optimally... "
let d = modularMultapilicativeInverse(e, totient)
echo "And were done!!!\n"

echo "p = " & $p & " q = " & $q
echo "d = " & $d

let elapsed = cpuTime() - start
echo "The whole process took *only* " & $elapsed & " seconds :)"

if output == "":
  output = readLineFromStdin("File to store private key (ENTER to not store it): ")

if output == "":
  quit()

let privateKey = %* {
  "n": n,
  "e": e,
  "d": d,
  "p": p,
  "q": q,
}

var
  f: File
if open(f, output):
  let proceed: string = readLineFromStdin("Are you sure you want to overwrite file \"" & output & "\" (Y or N): ")

  if normalize(proceed) != "y": #more like Y or anything besides Y
    if normalize(proceed) == "n":
      echo "Quitting"
      quit()
    else:
      echo "Unexpected Input: Quitting"
      quit()

try:
  writeFile(output, $privateKey)
except:
  echo "Unable to write to file \"" & output & "\""
