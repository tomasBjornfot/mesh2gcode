import numpy as np
import matplotlib.pyplot as plt
import time
import gcode_analyser as gan
import machining as m

"""
tg = m._connect()
for i in range(10):
    m._setGcodeValue(tg, "M3 S1000")
    print("Spindle on")
    time.sleep(5)
    m._setGcodeValue(tg, "M5")
    print("Spindle off")
    time.sleep(5)
m._disconnect(tg)
"""

def get_gcode_points(filename):
    with open(filename, 'r') as f:
        lines = f.readlines()
        points = []
        for line in lines:
            x = float(line.split(' ')[1][1:])
            y = float(line.split(' ')[2][1:])
            z = float(line.split(' ')[3][1:])
            points.append([x,y,z])
    return np.array(points)
def write_points_to_file(filename, points):
    with open(filename, 'w') as f:
        for i in range(len(points[:,0])):
            x = str(points[i, 0])
            y = str(points[i, 1])
            z = str(points[i, 2])
            f.write(x+' '+y+' '+z+'\n')
#--- MAIN ---#
data = get_gcode_points('cam/deck_surface.gc')
data[:, 1] = data[:, 1] - 801.95
write_points_to_file('out/deck.asc', data)
data = get_gcode_points('cam/bottom_surface.gc')
data[:, 1] = data[:, 1] - 801.95
data[:, 2] = -data[:, 2]
write_points_to_file('out/bottom.asc', data)
print('done')

"""
p, f = gan.data_from_gcode('cam/bottom.gc')
d = gan.p2p_distance(p)
t = gan.milling_time_segment(p, f)
index = []
for i in range(len(t)):
    if t[i] < 3.0*0.001:
        index.append(i)
print('done..')

p, f = gan.data_from_gcode('cam/deck.gc')
d = gan.p2p_distance(p)
t = gan.milling_time_segment(p, f)
index = []
for i in range(len(t)):
    if t[i] < 3.0*0.001:
        index.append(i)
print('done..')
"""
