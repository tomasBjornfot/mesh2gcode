import numpy as np
import matplotlib.pyplot as plt

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
    
# --- MAIN --- #
"""
p, f = data_from_gcode('cam/merge.gc')
print("Milling time:",milling_time(p,f))
"""
