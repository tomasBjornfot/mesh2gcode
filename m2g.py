# -*- coding: utf-8 -*-
import numpy as np
import matplotlib.pyplot as plt
import json
import makespiral
import gcode_analyser
import pdb


# --- readers ---
def read_settings():
    with open('settings.json', 'r') as f:
        data = json.load(f)
    return data
def read_matrices(side):
    if side == 'bottom':
        mx = read_jagged_matrix('out/bottom_mx.txt')
        my = read_jagged_matrix('out/bottom_my.txt')
        mz = read_jagged_matrix('out/bottom_mz.txt')
        zn = read_jagged_matrix('out/bottom_zn.txt')
    if side == 'deck':
        mx = read_jagged_matrix('out/deck_mx.txt')
        my = read_jagged_matrix('out/deck_my.txt')
        mz = read_jagged_matrix('out/deck_mz.txt')
        zn = read_jagged_matrix('out/deck_zn.txt')
    return mx, my, mz, zn
def merge_gcodefiles(path1, path2, path3):
    with open(path1, 'r') as f:
        gcode1 = f.readlines()
    with open(path2, 'r') as f:
        gcode2 = f.readlines()
    with open(path3, 'w') as f:
        gcode = gcode1 + gcode2
        for line in gcode:
            f.write(line)
# --- from go program ---
def read_jagged_matrix(path):
	data = []
	with open(path, 'r') as f:
		for line in f.readlines():
			data.append([float(d) for d in line.strip().split(' ')])
	return data	
def write_jagged_matrix(data, path):
    lines =[]
    for row in data:
        line = ''
        for item in row:
            line += str(item)+' '
        lines.append(line)
    with open(path, 'w') as f:
        [f.write(item+'\n') for item in lines]
# --- points <-> gcode ---
def points_to_gcode(points, feedrate, path):
	lines = []
	for i in range(len(points)):
		line = 'G1'
		line += ' X'+str(points[i,0])
		line += ' Y'+str(points[i,1])
		line += ' Z'+str(points[i,2])
		line += ' F'+str(feedrate[i])+'\n'
		lines.append(line)
	with open(path, 'w') as f:
		f.writelines(lines)
def points_from_gcode(path):
    points, feed = [], []
    with open(path, 'r') as f:
        for line in f.readlines():
            s = line.strip().split(' ')
            points.append([float(s[1][1:]), float(s[2][1:]), float(s[3][1:])])
            feed.append(float(s[4][1:]))
    return np.array(points), np.array(feed)
# --- analyser --- #
def point_distance(points):
	dist =  np.sqrt(np.sum(np.diff(points, axis=0)**2, axis=1))
	return dist, np.sum(dist)
# --- handling of handles --- #
def add_handles(mz, handleposition, handleheight, handlewidth):
    for index in range(len(handleposition)):
        center_pos = int(handleposition[index]*len(mz)+0.5)
        start_row = center_pos - handlewidth[index]
        end_row = center_pos + handlewidth[index]
        for row in range(start_row, end_row):
            for col in range(handleheight[index]):
                mz[row][col] = mz[row][handleheight[index]+1]
    return mz
# --- feedrate calculations --- #
def make_angle_deviation_z_matrix(zn):
    dz = []
    for row in range(len(zn)):
        dz_row = []
        for col in range(len(zn[row])-1):
            dev = np.abs(zn[row][col] - zn[row][col+1])
            dz_row.append(np.round(180/np.pi*np.sin(dev)))
        dz_row.append(0.0)
        dz.append(dz_row)
    return dz
def make_height_deviation_z_mat(mz, mateial_height):
    dz_height = []
    for row in range(len(mz)):
        dz_row = []
        for col in range(len(mz[row])):
            dz_row.append(np.round(mateial_height - mz[row][col]))
        
        dz_height.append(dz_row)
    return dz_height
def calculate_feedrate(mx, dz_angle, dz_height, max_feed, min_feed, halv_feed_a, half_feed_h):
    """
    calculates feedrate for all points in jagged matrices
    max_feed is the maximum feedrate 
    min_feed is the minimum feedrate 
    halv_feed_a is where the feedrate should be the half of 
    the max feed for the angulation
    halv_feed_a is where the feedrate should be the half of 
    max feed for the material depth
    """
    fr = []
    a_const = 1/(2*halv_feed_a)
    h_const = 1/(2*half_feed_h)
    for row in range(len(mx)):
        fr_row = []
        for col in range(len(mx[row])):
            
            # adapt to angle
            a_val = dz_angle[row][col]
            a_factor = a_val*a_const
            a_feed = a_factor*max_feed
            
            # adapt to milling depth
            h_val = dz_height[row][col]
            h_factor = h_val*h_const
            h_feed = h_factor*max_feed
            
            feed = max_feed - a_feed - h_feed
            if feed < min_feed:
                feed = min_feed
            fr_row.append(np.round(feed))
            
        fr.append(fr_row)
    return fr
# --- milling path and final gcode --- #
def add_start_point(points, feed, max_z, start_feed):
    start_point = points[0, :].tolist()
    start_point[2] = max_z
    points = points.tolist()
    points.insert(0, start_point) 
    feed = [start_feed]+feed.tolist()
    return np.array(points), np.array(feed)
def add_end_point(points, feed, max_z, end_feed):
    end_point = points[-1, :].tolist()
    end_point[2] = max_z
    points = points.tolist()
    points.append(end_point)
    feed = feed.tolist()
    feed.append(end_feed)
    return np.array(points), np.array(feed)

def make_milling_points(mx, my, mz, feed):
    gx, gy, gz, fr = [], [], [], []
    max_len_cols = np.max([len(m) for m in mx])
    for col in range(max_len_cols):
    # left side
        for row in range(len(mx)):
            if col<len(mx[row])-1:
                gx.append(mx[row][col])
                gy.append(my[row][col])
                gz.append(mz[row][col])
                fr.append(feed[row][col])
        # right side
        for row in range(len(mx)):
            row2 = len(mx)-row-1
            if col<len(mx[row2])-1:
                gx.append(-mx[row2][col])
                gy.append(my[row2][col])
                gz.append(mz[row2][col])
                fr.append(feed[row2][col])
    gx = list(reversed(gx))
    gy = list(reversed(gy))
    gz = list(reversed(gz))
    fr = list(reversed(fr))
    points = np.array([gx, gy, gz]).T
    return points, np.array(fr)
def calc(side):
    # read settings and matrices
    s = read_settings()
    mx, my, mz, zn = read_matrices(side)
    # add handles
    mz = add_handles(mz, s['HandlePosition'], s['HandleHeight'], s['HandleWidth'])
    # calculate feedrate for the "surface milling"
    dz_angle = make_angle_deviation_z_matrix(zn)
    max_height = s['BlockThickness']/2.0 + s['ToolRadius']
    dz_height = make_height_deviation_z_mat(mz, s['BlockThickness']/2.0)
    mf = calculate_feedrate(mx, dz_angle, dz_height, 2500, 700, 25.0, 80.0)
    # make points for the gcode on the surface
    points, feed = make_milling_points(mx, my, mz, mf)
    points, feed  = add_start_point(points, feed, s['HomingOffset'][2], 1500)
    points, feed  = add_end_point(points, feed, s['HomingOffset'][2], 1500)
    # write stuff
    points[:, 1] -= np.min(points[:, 1])
    points_to_gcode(points, feed, 'cam/'+side+'_surface.gc')
    write_jagged_matrix(dz_angle, 'out/'+side+'_dz_angle.txt')
    write_jagged_matrix(dz_height, 'out/'+side+'_dz_height.txt')
    write_jagged_matrix(mf, 'out/'+side+'_feedrate.txt')
    
    # spirals
    x_offset = s['Xres']/2
    feed = s['FeedrateStringer']
    step = s['StepStringer']
    if side == 'deck':
        makespiral.make_deck_spiral(x_offset, my, mz, step, feed, max_height)
        spiral, spiral_feed = points_from_gcode('cam/deck_spiral.gc')
        
    if side == 'bottom':
        makespiral.make_bottom_spirals(x_offset, my, mz, step, feed, max_height)
        path1 = 'cam/bottom_spiral_head.gc'
        path2 = 'cam/bottom_spiral_tail.gc'
        merge_gcodefiles(path1, path2, 'cam/bottom_spiral.gc')
        spiral, spiral_feed = points_from_gcode('cam/bottom_spiral.gc')

    merge_gcodefiles('cam/'+side+'_spiral.gc', 'cam/'+side+'_surface.gc', 'cam/'+side+'.gc')
def calculate():
    calc('deck')
    calc('bottom')
# --- MAIN --- #	
calculate()
