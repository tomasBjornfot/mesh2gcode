# -*- coding: utf-8 -*-
from flask import Flask, render_template, request, url_for, redirect, flash
from werkzeug.utils import secure_filename
import numpy as np
import json
import os
import time
import uuid
import machining as m
import m2g
import gcode_analyser as gan
import glob
import pdb

app = Flask(__name__)
UPLOAD_FOLDER = 'in'
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
app.secret_key = 'A0Zr98j/3yX R~XHH!jmN]LWX/,?RT'

# ----------- DEVELOPEMENT ----------- #
def _getPlotData(gcodefile=''):
    x = []; y = []; z = []
    if gcodefile:
        with open(gcodefile, 'r') as f:
            data = f.readlines()
        for d in data:
            x.append(float(d.split()[1][1:]))
            y.append(float(d.split()[2][1:]))
            z.append(float(d.split()[3][1:]))
    return [x, y, z]

"""
=============================
====  PUBLIC FUNCTIONS  ====
=============================
"""

dev_message = ''
plotfile = ''

@app.route('/', methods = ['GET','POST'])
def development():
    return render_template('index_ver0.html', bottom=[], deck=[], message=[])

@app.route('/stl', methods = ['GET', 'POST'])
def stl():
    if request.method == 'POST':
        f = request.files['file']
        fname = secure_filename(f.filename)
        fname_upload = os.path.join(UPLOAD_FOLDER, fname)
        f.save(fname_upload)
        # make deck and bottom gcode
        os.system('./mesh2gcode_ver2 '+fname_upload)
        block_thicknes = float(request.form['radio'])
        with open('settings.json', 'r') as f:
            data = json.loads(f.read())
        data['BlockThickness'] = block_thicknes
        with open('settings.json', 'w') as f:
            f.write(json.dumps(data, indent=4))
        m2g.calculate()
        bottom = _getPlotData('cam/bottom.gc')
        deck = _getPlotData('cam/deck.gc')
        time.sleep(0.1)
        # info
        dev_message = ['file: '+fname]
        p_bottom, f_bottom = gan.data_from_gcode('cam/bottom.gc')
        p_deck, f_deck = gan.data_from_gcode('cam/deck.gc')
        time_bottom = str(int(gan.milling_time(p_bottom, f_bottom)))
        time_deck = str(int(gan.milling_time(p_deck, f_deck)))
        dev_message.append('Milling time: '+time_bottom + '/' + time_deck+' minutes')
        dev_message.append('Width: '+str(2*int(np.max(p_bottom[:, 0])))+' mm')
        dev_message.append('Length: '+str(int(np.max(p_bottom[:, 1])))+' mm')
        dev_message.append('Height: '+str(2*int(np.max(p_bottom[:, 2])))+' mm')
        return render_template('index_ver0.html', bottom=bottom, deck=deck, message=dev_message)

@app.route('/tostart', methods = ['GET','POST'])
def tostart():
    dev_message = []
    bottom = _getPlotData('cam/bottom.gc')
    deck = _getPlotData('cam/deck.gc')
    try:
        m._moveToStartPosition()
        dev_message.append('At start position!')
    except Exception, e:
        dev_message.append(str(e))
    return render_template('index_ver0.html', bottom=bottom, deck=deck, message=dev_message)

@app.route('/homing', methods = ['GET','POST'])
def homing():
    dev_message = []
    bottom = _getPlotData('cam/bottom_merge.gc')
    deck = _getPlotData('cam/deck_merge.gc')
    try:
        m.homing()
        dev_message.append('At home position!');
    except Exception, e:
        dev_message.append(str(e))
    return render_template('index_ver0.html', bottom=bottom, deck=deck, message=dev_message)

@app.route('/milldeck', methods = ['GET','POST'])
def milldeck():
    dev_message = []
    try:
        m.millDeck()
        dev_message.append('Milling deck done!');
    except Exception, e:
        dev_message.append(str(e))
    return render_template('index_ver0.html', bottom=[], deck=[], message=dev_message)

@app.route('/millbottom', methods = ['GET','POST'])
def millbottom():
    dev_message = []
    try:
        m.millBottom()
        dev_message.append('Milling bottom done');
    except Exception, e:
        dev_message.append(str(e))
    return render_template('index_ver0.html', bottom=[], deck=[], message=dev_message)

# start the server
if __name__ == '__main__':
    app.run(debug=True)
