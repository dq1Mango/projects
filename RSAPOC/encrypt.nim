import os
import std/math
import std/rdstdin
import json
from std/strutils import parseInt, normalize

proc printHelp() = 
  echo """

Silly little RSA demonstration written in Nim by Thomas Ranney

  ./encrypt [publicKey] [options]

Options:
  -f: file        input file to encrypt (defaults to stdin)
  -o: file        file to write output to (defaults to stdout)

Note: if an input file is supplied without an output file, the input is overwritten
"""
  quit()


proc badUsage() = 
    echo "ERROR, bad usage \nsee encrypt --help for more details"
    quit()

proc getInput(file: string): int=
  try:
    let input: int = parseInt(open(file).readLine())
    return input
  except IOError:
    echo "ERROR: Could not open input file (does it exist?)"
    quit()
  except ValueError:
    echo "ERROR: Supplied data is not in integer form"
    quit()
  
type 
  Public = object 
    n: int 
    e: int

let args = commandLineParams()
let length = args.len

var #mhhhh very consistent variable names, types, and purposes
  input: int
  output: string = ""

if length == 0:
  badUsage()

elif length < 2:
  if args[0] == "--help":
    printHelp()
  
  try:
    input = parseInt(readLineFromStdin("Enter data (in integer form) to encrypt: "))
  except ValueError:
    echo "ERROR: supplied data is not in integer form"
    quit()

  output = "stdout" #lets hope no one names their output file "stdout"

elif length == 3:
  if args[1] == "-f":
    input = getInput(args[2])
    output = args[2]
  
  elif args[1] == "-o":#ok sue me theres only two options
    try:
      input = parseInt(readLineFromStdin("Enter data (in integer form) to encrypt: "))
    except ValueError:
      echo "ERROR: supplied data is not in integer form"
      quit()

    output = args[2]

  else:
    badUsage()

elif length == 5:
  if args[1] == "-f" and args[3] == "-o":
    input = getInput(args[2])
    output = args[4]

  elif args[3] == "-f" and args[1] == "-o":
    input = getInput(args[4])
    output = args[2]
 

  else:
    badUsage()
   
else:
  badUsage()

var key: Public
try:
  key = to(parseJson(readFile(args[0])), Public)
except IOError:
  echo "ERROR: Cannot open public key file (does it exist)"
  quit()
except ValueError:
  echo "ERROR: Public key of incorrect format (run ./generateRSAKeysInsteadOfEnglish to generate a valid public key)"
  quit()

let n = key.n
let e = key.e

echo "public key:", key
echo "data:", input

proc modularMathILearnedFromKeyser(b: int, e: int, m: int): int =
  result = b
  for i in 2..e:
    result = b * result mod m

let cypher: int = modularMathILearnedFromKeyser(input, e, n)

if output == "stdout":
  echo "Encrypted data: " & $cypher
  quit()

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
  writeFile(output, $cypher)
except:
  echo "Unable to write to file \"" & output & "\""
  echo "Heres your encrypted data anyway: " & $cypher


