import copy
import sys
import os
from time import sleep
import pickle

# lol the classic
sys.setrecursionlimit(100000)

def emptyBoard():
    return [[0 for i in range(4)] for i in range(5)]

board = emptyBoard()


def dispBoard(baord):
    for row in baord:
        print("[", end="")
        for num in row:
            print('\033[9' + str(num) + 'm' + str(num), end = ", ")

        print('\033[0m ', end = "")
        print("]")

rawPieces = [
    (2, 2),
    (1, 2)] + \
    [(2, 1) for i in range(4)] + \
    [(1, 1) for i in range(4)]

graph = {}
graph[0] = {"adjacents": [], "coords": [(0, 1), (2, 1), (0, 0), (2, 0), (0, 3), (2, 3), (4, 0), (3, 1), (4, 3), (3, 2)]}

nextIndex = 1

winners = []

def boardFromCoords(coords):
        
    board = emptyBoard()

    for index in range(len(coords)):
        piece = rawPieces[index]
        for i in range(piece[0]):
            for j in range(piece[1]):
                board[coords[index][0] + i][coords[index][1] + j] = index + 1
    return board


def validMove(piece, oldPos, change, board):
    
    pos = tuple(map(sum, zip(oldPos, change))) # you gotta love python sometimes
    
    # i wish pitt offered a class on collision logic
    if pos[0] < 0 or pos[1] < 0 or pos[0] + piece[0] - 1 > 4 or pos[1] + piece[1] - 1 > 3:
        return False

    # wish i knew a better way to implement this logic
    match change:
        case (1, 0):
            for j in range(piece[1]):
                if board[pos[0] + piece[0] - 1][pos[1] + j] != 0:
                    return False
        case (-1, 0):
            for j in range(piece[1]):
                if board[pos[0]][pos[1] + j] != 0:
                    return False
        case (0, 1):
            for i in range(piece[0]):
                if board[pos[0] + i][pos[1] + piece[1] - 1] != 0:
                    return False
        case (0, -1):
            for i in range(piece[0]):
                if board[pos[0] + i][pos[1]] != 0:
                    return False

    return True

def coordsEqual(coordsOne, coordsTwo):

    if coordsOne[0] != coordsTwo[0]:
        return False
    if coordsOne[1] != coordsTwo[1]:
        return False

    for i in range(4):
        if coordsTwo[2 + i] not in coordsOne[2:2+4]:
            return False
    for i in range(4):
        if coordsTwo[6 + i] not in coordsOne[6:6+4]:
            return False

    return True


def nodeExists(coords):

    for index, node in graph.items():
        if coordsEqual(coords, node["coords"]):
            return index
    return -1
        
def addNode(node, index, change, oldId):
    coords = copy.deepcopy(node["coords"])
    
    # this really annoying amalgomation happedns twice!!! which bothers me
    coords[index] = tuple(map(sum, zip(coords[index], change))) # it adds tuples
    
    edge = nodeExists(coords)
    if edge != -1 :
        graph[oldId]["adjacents"].append(edge)
        graph[edge]["adjacents"].append(oldId)
        return
    
    newNode = {"adjacents": [oldId], "coords": coords}
        
    global nextIndex

    node["adjacents"] += [nextIndex]
    graph[nextIndex] = newNode
    nextIndex += 1

    detectWin(coords)

    addAdjacents(nextIndex - 1)


def addAdjacents(nodeId):
        
    print(nodeId)
    if nodeId > 1000:
        print("ended the bad way")
        return

    node = graph[nodeId]
    coords = node["coords"]
    board = boardFromCoords(coords)

    for i in range(len(rawPieces)):
        piece = rawPieces[i]

        moves = [(1, 0), (-1, 0), (0, 1), (0, -1)]
        for move in moves:
            if validMove(piece, coords[i], move, board):
                addNode(node, i, move, nodeId)

def detectWin(coords):
    if coords[0] == (3, 1):
        global nextIndex
        global winners
        winners += [nextIndex - 1]
        return

        print("we have a winner:", nextIndex - 1)
        print(coords)
        dispBoard(boardFromCoords(coords))
        holyGrail = findPath(nextIndex - 1, [])
        print(len(holyGrail))
        #tracePath(holyGrail)
        os._exit(0)

# dont print the thousands of lines of recursive stack trace
try:
    addAdjacents(0)
except KeyboardInterrupt:
    os._exit(1)

with open('graph.pkl', 'wb') as f:
    pickle.dump(graph, f)
