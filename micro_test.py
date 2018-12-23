import numpy as np
import matplotlib.pyplot as plt
import time
import gcode_analyser as gan
import machining as m

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
