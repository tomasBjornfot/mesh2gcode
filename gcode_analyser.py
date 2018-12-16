import numpy as np
import matplotlib.pyplot as plt
import pdb

def data_from_gcode(path):
    """
    get the points and feedrate from a gcode file
    """
    points = []
    feedrate = []
    with open(path, 'r') as f:
        for line in f.readlines():
            s = line.strip().split(' ')
            points.append([float(s[1][1:]), float(s[2][1:]), float(s[3][1:])])
            feedrate.append(float(s[4][1:]))
    return np.array(points), np.array(feedrate)
def p2p_distance(points):
    """
    point to point distance
    """
    dist =  np.sqrt(np.sum(np.diff(points, axis=0)**2, axis=1))
    return dist

def milling_time(points, feedrate):
    dist = p2p_distance(points)
    m = 0.0
    for i in range(len(dist)):
        m += dist[i]/feedrate[i+1]
    return m
def milling_time_segment(points, feedrate):
    """
    calculates the time i takes to travel to the next gcode point.
    This is important since Tinyg reports an error (202) if less
    then MIN_SEGMENT_USEC.
    """
    dist = p2p_distance(points)
    t = []
    for i in range(len(dist)):
        t.append(60*dist[i]/feedrate[i+1])
    return t
def is_p2p_time_ok(path, min_time):
    points, feedrate = data_from_gcode(path)
    t = milling_time_segment(points, feedrate)
    if np.min(t) < min_time:
        return False
    return True
# --- MAIN --- #
"""
p, f = data_from_gcode('cam/merge.gc')
print("Milling time:",milling_time(p,f))
"""
