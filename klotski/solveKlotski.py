from graphKlotski import boardFromCoords
import pickle
from time import sleep

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

def findPath(nodeId, path):
    path.insert(0, nodeId)

    if nodeId == 0:
        return path
    
    deadEnd = True
    for edge in graph[nodeId]["adjacents"]:

        if edge in path:
            continue
        deadEnd = False

        if findPath(edge, path) != []:
            return path

        path.pop()
    
    if deadEnd:
        return [] 
def tracePath(path):
    for node in path:
        print(node, " -- ")
        print(graph[node]["adjacents"])
        dispBoard(boardFromCoords(graph[node]["coords"]))
        print()

# shhhh i cant be bothered to write solutions indexes to file
path = findPath(965, [])

for i, node in enumerate(path[-1:]):
    for edge in graph[node]["adjacents"]:
        if edge in path and path[i + 1] != edge:
            print("found a shorter route")
            del path[i + 1: path.index(edge)] 

tracePath(path)
print(len(path))
