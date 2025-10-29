from random import randint
import copy
import curses
import os
from time import sleep

os.system('stty sane')

size = int(input('what size board would you like? \n'))

stdscr = curses.initscr()
begin_x = 6; begin_y = 3
height, width = stdscr.getmaxyx()
win = curses.newwin(height - 3, width - 6, begin_y, begin_x)
curses.curs_set(0)
curses.cbreak()
curses.noecho()
win.keypad(True)

actualBoard = [[0 for y in range(size)] for x in range(size)]

score = 0
merges = 0
turn = 0

start1 = [randint(0, size - 1), randint(0, size - 1)]

while True:
    start2 = [randint(0, size - 1), randint(0, size - 1)]
    if start2 != start1:
        break

actualBoard[start1[0]][start1[1]], actualBoard[start2[0]][start2[1]] = 2, 2


tileWeight = 1
cornerWeight = 10
monotomicWieght = 4
adjWeight = 9
def rateBoard(board):
    boardScore = 0
    biggest = [[-1]]
    y = 0
    for i in board:
        x = 0
        for j in i:
            try:
                if board[y][x + 1] < board[y][x]: boardScore += board[y][x + 1] * monotomicWieght
                elif board[y][x + 1] == board[y][x]: boardScore += board[y][x] * adjWeight
                else: boardScore -= board[y][x + 1] * monotomicWieght

                if board[y + 1][x] < board[y][x]: boardScore += board[y + 1][x] * monotomicWieght
                elif board[y + 1][x] == board[y][x]: boardScore += board[y + 1][x] * adjWeight
                else: boardScore -= board[y + 1][x] * monotomicWieght
            except: pass

            if j >= biggest[0][0]:
                if j == biggest[0][0]:
                    biggest.append([j, y, x])
                else: biggest[0] = [j, y, x]
            if j == 0: boardScore += tileWeight
            x += 1
        y += 1

    for i in biggest:
        if i[1] == 0 and i[2] == 0:
            boardScore += i[0] * cornerWeight
            break

    return boardScore

moves = {}
def updateMoves(board1):
    check_board = copy.deepcopy(board1)
    if moveRight(False, board1) != check_board: moves['right'] = rateBoard(moveRight(False, board1))
    else:
        try: moves.pop('right')
        except: pass
    if moveLeft(False, board1) != check_board: moves['left'] = rateBoard(moveLeft(False, board1))
    else:
        try: moves.pop('left')
        except: pass
    if moveUp(False, board1) != check_board: moves['up'] = rateBoard(moveUp(False, board1))
    else:
        try: moves.pop('up')
        except: pass
    if moveDown(False, board1) != check_board: moves['down'] = rateBoard(moveDown(False, board1))
    else:
        try: moves.pop('down')
        except: pass

def dispBoard():
    win.clear()
    y = 0
    while y < size:
        x = 0
        while x < size:
            for i in range(1, 11):
                if actualBoard[y][x] == 2 ** i:
                    win.addstr(y * 3, x * 6, str(actualBoard[y][x]))
                    break
            else :
                win.addstr(y * 3, x * 6, str(actualBoard[y][x]))
            x += 1
        win.addstr("\n")
        y += 1
    win.refresh()

    win.addstr('\nscore : ' + str(score) + '\n')
    #win.addstr('\ncalculated board score : ' + str(rateBoard(board)))

# moveRight takes the actual borad and returns a NEW board having applied the move
def moveRight(doUpdate, board_in):
    board = copy.deepcopy(board_in)

    for i in range(size - 1):
        y = 0
        while y < size:
            x = size - 1
            while x > 0:
                if board[y][x] == 0:
                    board[y][x] = board[y][x - 1]
                    board[y][x - 1] = 0
                x = x - 1
            y = y + 1

    y = 0
    while y < size:
        x = size - 1
        while x > 0:
            if board[y][x] == board[y][x - 1] and board[y][x] != 0:
                board[y][x] *= 2
                board[y][x - 1] = 0
                if doUpdate == True: update(x, y)
            x -= 1
        y += 1

    y = 0
    while y < size:
        x = size - 1
        while x > 0:
            if board[y][x] == 0:
                board[y][x] = board[y][x - 1]
                board[y][x - 1] = 0
            x = x - 1
        y = y + 1

    return board

def moveUp(doUpdate, board_in):
    board = copy.deepcopy(board_in)

    for i in range(size - 1):
        x = 0
        while x < size:
            y = 0
            while y < size -1:
                if board[y][x] == 0:
                    board[y][x] = board[y + 1][x]
                    board[y + 1][x] = 0
                y = y + 1
            x += 1

    x = 0
    while x < size:
        y = 0
        while y < size - 1:
            if board[y][x] == board[y + 1][x] and board[y][x] != 0:
                board[y][x] *= 2
                board[y + 1][x] = 0
                if doUpdate == True: update(x, y)
            y += 1
        x += 1

    x = 0
    while x < size:
        y = 0
        while y < size -1:
            if board[y][x] == 0:
                board[y][x] = board[y + 1][x]
                board[y + 1][x] = 0
            y = y + 1
        x += 1
    return board

def moveLeft(doUpdate, board_in):
    board = copy.deepcopy(board_in)

    for i in range(size - 1):
        y = 0
        while y < size:
            x = 0
            while x < size - 1:
                if board[y][x] == 0:
                    board[y][x] = board[y][x + 1]
                    board[y][x + 1] = 0
                x = x + 1
            y = y + 1
    y = 0
    while y < size:
        x = 0
        while x < size - 1:
            if board[y][x] == board[y][x + 1] and board[y][x] != 0:
                board[y][x] *= 2
                board[y][x + 1] = 0
                if doUpdate == True: update(x, y)
            x += 1
        y += 1

    y = 0
    while y < size:
        x = 0
        while x < size - 1:
            if board[y][x] == 0:
                board[y][x] = board[y][x + 1]
                board[y][x + 1] = 0
            x = x + 1
        y = y + 1
    return board

def moveDown(doUpdate, board_in):
    board = copy.deepcopy(board_in)

    for i in range(size - 1):
        x = 0
        while x < size:
            y = size - 1
            while y > 0:
                if board[y][x] == 0:
                    board[y][x] = board[y - 1][x]
                    board[y - 1][x] = 0
                y = y - 1
            x = x + 1

    x = 0
    while x < size:
        y = size - 1
        while y > 0:
            if board[y][x] == board[y - 1][x] and board[y][x] != 0:
                board[y][x] *= 2
                board[y - 1][x] = 0
                if doUpdate == True: update(x, y)
            y -= 1
        x += 1
    x = 0
    while x < size:
        y = size - 1
        while y > 0:
            if board[y][x] == 0:
                board[y][x] = board[y - 1][x]
                board[y - 1][x] = 0
            y = y - 1
        x = x + 1

    return board

def update(x, y):
    global score
    global merges
    score += actualBoard[y][x]
    merges += 1

dispBoard()
while True:

    sleep(0.01)
    updateMoves(actualBoard)

    try: bestMove = list(list(moves.items())[0])
    except: break
    for key, value in moves.items():
        if value > bestMove[1]:
            bestMove[0] = key
            bestMove[1] = value

    if bestMove[0] == 'right': actualBoard = moveRight(True, actualBoard)
    elif bestMove[0] == 'left': actualBoard = moveLeft(True, actualBoard)
    elif bestMove[0] == 'up': actualBoard = moveUp(True, actualBoard)
    elif bestMove[0] == 'down': actualBoard = moveDown(True, actualBoard)

    while True:
            newTile = [randint(0, size - 1), randint(0, size - 1)]
            if actualBoard[newTile[0]][newTile[1]] == 0:
                number = randint(1, 10)
                if number == 10: actualBoard[newTile[0]][newTile[1]] = 4
                else: actualBoard[newTile[0]][newTile[1]] = 2
                break

    dispBoard()



win.clear()
curses.nocbreak()
win.keypad(False)
curses.echo()
curses.curs_set(1)
curses.endwin()

os.system('clear')


print('\n')
for i in actualBoard:
    for j in i: print(j, ' ', end = '')
    print('\n')
print('score : ', score)
print('_____.___.                   .____                     __    ')
print('\__  |   |  ____   __ __     |    |     ____   _______/  |_  ')
print(' /   |   | /  _ \ |  |  \    |    |    /  _ \ /  ___/\   __\ ')
print(' \____   |(  <_> )|  |  /    |    |___(  <_> )\___ \  |  |   ')
print(' / ______| \____/ |____/     |_______ \\\____//____  > |__|   ')
print(' \/                                  \/           \/         ')
