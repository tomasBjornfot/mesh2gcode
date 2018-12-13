# -*- coding: utf-8 -*-
import numpy as np
import matplotlib.pyplot as plt
import pdb, json

stop = pdb.set_trace

###########################################
##########   PRIVATE FUNCTIONS   ##########
###########################################
def split_head_tail(spiral_org):
    # spilt top and tail to two sets
    index = np.argmax(spiral_org[2])
    head = [np.copy(spiral_org[i][(index-1):]) for i in range(3)]
    tail = [np.copy(spiral_org[i][:index]) for i in range(3)]    
    return head, tail, index
    
def get_patches(z, step, max_height):
    patches = []
    for i in range(100):
        z = z + step
        for i in range(len(z)):
            if z[i] >= max_height:
                z[i] = max_height
        if np.min(z) == max_height:
            break
        patches.append(z)
    return patches[::-1]

def find_magic_index(z, max_height, what='head'):
    if what=='head':
        for i in range(len(z)):
            if z[i] != max_height:
                return i
    if what=='tail':
        for i in range(len(z)):
            if z[len(z)-i-1] != max_height:
                return i
    # return -1 if all values are above max_height
    return -1

def points_to_gcode(points, fr):
    lines = []
    for p in points:
        line = 'G1 X'
        line += str(p[0])+' Y'
        line += str(p[1])+' Z'
        line += str(p[2])+' F'+str(fr)
        lines.append(line)    
    return lines
    
def write_gcode(myfile, lines):
    with open('settings.json', 'r') as f:
        maxz = json.loads(f.read())['HomingOffset'][2]
    # adding a point above the first point
    line = lines[0].split()
    line[3] = 'Z'+str(maxz)
    first_line = line[0]+' '+line[1]+' '+line[2]+' '+line[3]+' '+line[4]
    lines = [first_line] + lines
    # adding a point above the last point
    line = lines[-1].split()
    line[3] = 'Z'+str(maxz)
    last_line = line[0]+' '+line[1]+' '+line[2]+' '+line[3]+' '+line[4]
    lines = lines + [last_line]
    with open(myfile, 'w') as f:
        for line in lines:
            f.write(line+'\n')


#######################################
##########  PUBLIC FUNCTIONS ##########
#######################################
def make_deck_spiral(x_offset, y_offset, my, mz, step, feedrate, max_height):
    """
    x_offset, the distance between the board center and milling patch in x direction and mm
    my, jagged matrix from golang code
    mz,jagged matrix from golang code
    step, the inclination step per spiral
    feedrate, the feedrate
    max_height, the max height (normally BlockThickness/2)
    """
    # gets the patch
    patch = []
    x = -x_offset*np.ones(len(my))
    y = np.array([row[-1] for row in my])
    z = np.array([row[-1] for row in mz])
    patch = [x, y, z]

    # makes a set of spirals that offsets in z by 'step'
    z = get_patches(patch[2], step, max_height)

    # get the magic indices
    lower_magic_index = []
    upper_magic_index = []

    for i in range(len(z)):
        #lower index
        for j in range(len(z[i])):
            if z[i][j] != max_height:
                lower_magic_index.append(j)
                break
        #upper index
        for j in range(len(z[i])):
            jj = len(z[i]) - j - 1
            if z[i][jj] != max_height:
                upper_magic_index.append(jj)
                break
    # remove points at max height
    # create a spiral for each height level
    spirals = []
    for i in range(len(z)):
        _start = lower_magic_index[i]
        _stop = upper_magic_index[i]
        _x = patch[0][_start:_stop]
        _y = patch[1][_start:_stop]
        _z = z[i][_start:_stop]
        spirals.append([_x, _y, _z])

    deck_gcode = []
    for s in spirals:
        # forward on x positive side
        for i in range(len(s[0])):
            deck_gcode.append([s[0][i], s[1][i], s[2][i]])
        # backward on x negative side
        for i in range(len(s[0])):
            ii = len(s[0]) - 1 - i
            deck_gcode.append([-s[0][ii], s[1][ii], s[2][ii]])
    # moving the y-values to machine domain
    for dc in deck_gcode:
        dc[1] -= y_offset
    lines = points_to_gcode(deck_gcode, feedrate)
    write_gcode('cam/deck_spiral.gc', lines)
def make_bottom_spirals(x_offset, y_offset, my, mz, step, feedrate, max_height):
    # gets the patch
    patch = []
    x = -x_offset*np.ones(len(my))
    y = np.array([row[-1] for row in my])
    z = np.array([row[-1] for row in mz])
    patch = [x, y, z]
    # split the data to head and tail
    head, tail, split_index = split_head_tail(patch)
    # makes a set of spirals that offsets in z by 'step'
    head_z = get_patches(head[2], step, max_height)
    tail_z = get_patches(tail[2], step, max_height)
    # removing all data points that was reached the max_height i z
    ## head
    magic_index_head = [find_magic_index(z, max_height, 'head') for z in head_z]
    head = []
    for i in range(len(head_z)):
        if magic_index_head[i] == -1 :
            break
        x = patch[0][(split_index - 1 + magic_index_head[i]):]
        y = patch[1][(split_index - 1 + magic_index_head[i]):]
        z = head_z[i][magic_index_head[i]:]
        head.append([x,y,z])
    ## tail
    magic_index_tail = [find_magic_index(z, max_height, 'tail') for z in tail_z]
    tail = []
    for i in range(len(tail_z)):
        if magic_index_tail[i] == -1 :
            break
        x = patch[0][:(split_index - 1 - magic_index_tail[i])]
        y = patch[1][:(split_index - 1 - magic_index_tail[i])]
        z = tail_z[i][:(split_index - 1 - magic_index_tail[i])]
        tail.append([x,y,z])
    # make milling patch
    ## head
    head_gcode = []
    for h in head:
        # forward on x positive side
        for i in range(len(h[0])):
            head_gcode.append([h[0][i], h[1][i], h[2][i]])
        # backward on x negative side
        for i in range(len(h[0])):
            ii = len(h[0]) - 1 - i
            head_gcode.append([-h[0][ii], h[1][ii], h[2][ii]])
    # moving the y-values to machine domain
    for hc in head_gcode:
        hc[1] -= y_offset
    # head to gcode file
    lines = points_to_gcode(head_gcode, feedrate)
    write_gcode('cam/bottom_spiral_head.gc', lines)

    ## tail
    tail_gcode = []
    for t in tail:
        # forward on x positive side
        for i in range(len(t[0])):
            ii = len(t[0]) - 1 - i
            tail_gcode.append([t[0][ii], t[1][ii], t[2][ii]])
        # backward on x negative side
        for i in range(len(t[0])):
            tail_gcode.append([-t[0][i], t[1][i], t[2][i]])
    # moving the y-values to machine domain
    for tc in tail_gcode:
        tc[1] -= y_offset
    lines = points_to_gcode(tail_gcode, feedrate)
    write_gcode('cam/bottom_spiral_tail.gc', lines)

