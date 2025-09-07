import pickle
from time import sleep
import sys
import copy
import subprocess

sys.setrecursionlimit(100000)

rawPieces = [
    (2, 2),
    (1, 2)] + \
    [(2, 1) for i in range(4)] + \
    [(1, 1) for i in range(4)]

def emptyBoard():
    return [[0 for i in range(4)] for i in range(5)]

def boardFromCoords(coords):
        
    board = emptyBoard()

    for index in range(len(coords)):
        piece = rawPieces[index]
        for i in range(piece[0]):
            for j in range(piece[1]):
                board[coords[index][0] + i][coords[index][1] + j] = index + 1
    return board

def dispBoard(baord):
    for row in baord:
        print("[", end="")
        for num in row:
            print('\033[9' + str(num) + 'm' + str(num), end = ", ")

        print('\033[0m ', end = "")
        print("]")

graph = {}

with open('graph.pkl', 'rb') as f:
    graph = pickle.load(f)

def trimPath(path):
    i = 0
    while i < len(path) - 1:
        node = path[i]
        #print(i)
        for edge in graph[node]["adjacents"]:
            if edge in path[i + 1:]:
                #print("found a shorter route")
                del path[i + 1: path.index(edge)] 
        i += 1
    return path

def findPath(nodeId, path, paths):
    path.insert(0, nodeId)

    if nodeId == 0:
        path = trimPath(path)
        if path not in paths:
            print("found a path of length:", len(path))
            paths.append(copy.deepcopy(path)) 
        path.pop()
        return
    
    deadEnd = True
    for edge in graph[nodeId]["adjacents"]:

        if edge in path:
            continue
        deadEnd = False

        findPath(edge, path, paths)

    path.pop()

def tracePath(path):
    for node in path:
        print(node, " -- ")
        print(graph[node]["adjacents"])
        dispBoard(boardFromCoords(graph[node]["coords"]))
        print()

allThePaths = []

# shhhh i cant be bothered to write solutions indexes to a file
try:
    findPath(965, [], allThePaths)
except KeyboardInterrupt:
    pass
except Exception as e:
    print( e)

sorted(allThePaths, key=lambda x: len(x))
bestPath = allThePaths[0]
tracePath(bestPath)
