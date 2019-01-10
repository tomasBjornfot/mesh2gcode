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
    """
    Reads the gcode data from file. Use this function to update the plot in webform.
    Args:
        gcodefile (string): Optional. The path to the gcode file
    Returns:
        list: The points in the gcode file
    """
    x = []; y = []; z = []
    if gcodefile:
        with open(gcodefile, 'r') as f:
            data = f.readlines()
        for d in data:
            x.append(float(d.split()[1][1:]))
            y.append(float(d.split()[2][1:]))
            z.append(float(d.split()[3][1:]))
    return [x, y, z]

def gcode_info():
    mess = []
    # checks is gcode files produce error 202:
    b_ok = gan.is_p2p_time_ok('cam/bottom.gc', 3.0*0.001)
    d_ok = gan.is_p2p_time_ok('cam/deck.gc', 3.0*0.001)
    if b_ok == False or d_ok == False:
        mess.append('Warning: files can cause error: 202')
    p_bottom, f_bottom = gan.data_from_gcode('cam/bottom.gc')
    p_deck, f_deck = gan.data_from_gcode('cam/deck.gc')
    time_bottom = str(int(gan.milling_time(p_bottom, f_bottom)))
    time_deck = str(int(gan.milling_time(p_deck, f_deck)))
    mess.append('Milling time: '+time_bottom + '/' + time_deck+' minutes')
    mess.append('Width: '+str(2*int(np.max(p_bottom[:, 0])))+' mm')
    mess.append('Length: '+str(int(np.max(p_bottom[:, 1])))+' mm')
    mess.append('Height: '+str(2*int(np.max(p_bottom[1:-1, 2])))+' mm')
    return mess

"""
=============================
====  PUBLIC FUNCTIONS  ====
=============================
"""

#dev_message = '' # Why is this value defined here? Remove?
#plotfile = '' # Only exist here! Remove?

@app.route('/', methods = ['GET','POST'])
def index():
    """
    Gets the template.
    Args:
        None
    Returns:
        ? : The HTML of the template. 
    """
    return render_template('index_ver0.html', bottom=[], deck=[], message=[])

@app.route('/stl', methods = ['GET', 'POST'])
def stl():
    """
    The function is triggered when the calculate button is pushed in the HTML.
    It calculates the gcode files for an stl file. It also gives the
    properties of the gcode, i.e. milling time and size
    Args:
        None
    Returns:
        ? : The HTML of the template
    """
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
            f.write(json.dumps(data, indent=4, sort_keys=True))
        m2g.calculate()
        m2g.split_gcodefiles()
        mess = ['File: '+fname]
        mess = mess + gcode_info()
        bottom = _getPlotData('cam/bottom.gc')
        deck = _getPlotData('cam/deck.gc')
        return render_template('index_ver0.html', bottom=bottom, deck=deck, message=mess)

@app.route('/tostart', methods = ['GET','POST'])
def tostart():
    """
    Moves the milling machine to the start point. This function
    shall not be used in the final relese.
    Args:
        None
    Returns:
        ? : The HTML of the template
    """
    mess = gcode_info()
    bottom = _getPlotData('cam/bottom.gc')
    deck = _getPlotData('cam/deck.gc')
    try:
        m._moveToStartPosition()
        mess.append('At start position!')
    except Exception as e:
        mess.append(str(e))
    return render_template('index_ver0.html', bottom=bottom, deck=deck, message=mess)

@app.route('/homing', methods = ['GET','POST'])
def homing():
    """
    Makes a homing of the machine
    Args:
        None
    Returns:
        ? : The HTML of the template
    """
    mess = gcode_info()
    bottom = _getPlotData('cam/bottom.gc')
    deck = _getPlotData('cam/deck.gc')
    try:
        m.homing()
    except Exception as e:
        mess.append(str(e))
    return render_template('index_ver0.html', bottom=bottom, deck=deck, message=mess)

@app.route('/milldeck', methods = ['GET','POST'])
def milldeck():
    """
    Performs a milling of the deck of the blank.
    Args:
        None
    Returns:
        ? : The HTML of the template
    """
    mess = gcode_info()
    try:
        mess.append(m.millDeck())
    except Exception as e:
        mess.append(str(e))
    return render_template('index_ver0.html', bottom=[], deck=[], message=mess)

@app.route('/millbottom', methods = ['GET','POST'])
def millbottom():
    """
    Performs a milling of the bottom of the blank.
    Args:
        None
    Returns:
        ? : The HTML of the template
    """
    mess = gcode_info()
    try:
        mess.append(m.millBottom())
    except Exception as e:
        mess.append(str(e))
    return render_template('index_ver0.html', bottom=[], deck=[], message=mess)

# start the server
if __name__ == '__main__':
    app.run(debug=True)
