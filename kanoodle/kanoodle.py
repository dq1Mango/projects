import random
from colorist import hex
import copy

#board = [[0 for i in range(11)] for i in range(5)] 

#incase u forgor 💀, u set the array to 1 where there is an initla piece (great ui ik)
OGboard = [[0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0],
         [0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0],
         [0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0],
         [0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0],
         [0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0]]

#u also have to comment out the pieces that were inital
startingPieces = {
        #"plus": [[0, 1, 0], [1, 1, 1], [0, 1, 0]],
        "lyreHate": [[1, 1, 0, 0], [0, 1, 1, 1]],
        "diagonal": [[1, 1, 0], [0, 1, 1], [0, 0, 1]], 
        "u": [[1, 0, 1], [1, 1, 1]],
        "lineButNot": [[1, 1, 1, 1,], [0, 0, 1, 0]],
        #"squareBotNot": [[1, 0], [1, 1,], [1, 1]],
        "angle": [[1, 1, 1,], [0, 0, 1], [0, 0, 1]],
        "square": [[1, 1], [1, 1]],
        "bigL": [[1, 1, 1, 1],[0, 0, 0, 1]],
        "smallL": [[1, 1, 1], [0, 0, 1]],
        "line": [[1, 1, 1, 1]],
        "corner": [[1, 1],[0, 1]]
        }
pieces = {}

colorMap = {
        "plus": "#C0C0C0",
        "lyreHate": "#006600",
        "diagonal": "#FF66B2", 
        "u": "#ffff00",
        "lineButNot": "#ffcce5",
        "squareBotNot": "#ff0000",
        "angle": "#00ffff",
        "square": "#66ffb2",
        "bigL": "#0000cc",
        "smallL": "#ff9933",
        "line": "#7f00ff",
        "corner": "#c0c0c0"
        }

def rotate(shape):
    newShape = []
    for i in range(len(shape[0])):
        newShape.append([])
        for j in range(len(shape) - 1, -1, -1):
            newShape[i].append(shape[j][i])

    return newShape

def flip(shape):
    newShape = [[0 for i in range(len(shape[0]))] for i in range(len(shape))]
    for i in range(len(shape)):
        for j in range(len(shape[i])):
            newShape[i][len(shape[i]) - j - 1] = shape[i][j]

    return newShape

i = 2
for piece in startingPieces:
    pieces[i] = []
    pieces[i].append(startingPieces[piece])

    testShape = rotate(startingPieces[piece])
    if testShape not in pieces[i]: pieces[i].append(testShape)

    testShape = rotate(testShape)
    if testShape not in pieces[i]: 
        pieces[i].append(testShape)
        pieces[i].append(rotate(testShape))
        
        testShape = flip(testShape)
        if testShape not in pieces[i]:
            for j in range(4):
                pieces[i].append(testShape)
                testShape = rotate(testShape)

    i += 1
def checkFit(boardButSpelledDifferentlyAsToNotMutateTheGlobalVariable, piece, number, row, col):
    for y in range(len(piece)):
        for x in range(len(piece[0])):
            if (boardButSpelledDifferentlyAsToNotMutateTheGlobalVariable[row + y][col + x] != 0 and piece[y][x] == 1): return False
            else: 
                #print("changed baord at x:", col + x, "y:", row + y)
                boardButSpelledDifferentlyAsToNotMutateTheGlobalVariable[row + y][col + x] += piece[y][x] * number

    return boardButSpelledDifferentlyAsToNotMutateTheGlobalVariable

def solve(board, number):
    print(number)
    for orientation in pieces[number]:
        for row in range(len(board) - len(orientation) + 1):
            for collum in range(len(board[0]) - len(orientation[0]) + 1):
                branchBoard = checkFit(copy.deepcopy(board), orientation, number, row, collum)
                if (branchBoard != False):
                    if (number == len(pieces) + 1): 
                        return branchBoard

                    branch = solve(branchBoard, number + 1)
                    if branch != False: return branch
    
    if (number == 2):
        print("i died :(", number)
        printBoard(branchBoard)
    return False

def printBoard(aBoard):
    if aBoard == False: print(aBoard)
    else: 
        for row in aBoard: print(row)

solved = solve(OGboard, 2)

print(solved)
