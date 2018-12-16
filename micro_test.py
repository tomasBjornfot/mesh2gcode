import numpy as np
import matplotlib.pyplot as plt
import gcode_analyser as gan

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