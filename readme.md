# Användarinstruktion:
Editera JSON filen så den passar materialet.
Skapa fräsfiler. En nära den rörliga infästningen och en nära den fasta infästningen
Fräs båda filerna, vänd materialet och fräs samma filer på andra sidan
Mät tjockleken på det utfrästa materialet: 

Felet i höjd (z) är HeightOffset[2] minus den uppmätta tjockleken, delat i två. 
Felet i x är halva skilladen mellan vertikala planen x och x-.

Addera felet med HeightOffset[2] gör en homing, för en ny fräsnng och mät.  

Parametrar i json filen:
ToolRadius, radien på fräsverktyget
MaxHeight, det högsta tillåtena z värdet. Samma som HomingOffset[2]
Xpos, positionen i xled som avstånd från centrum
Ypos, positionen i yled 
MaterialThickness, tjockleken på materialet
FilePrefix, om man vill ge ett mer specifikt namn på g-kod filen

För att köra en fil som gör ett plan i centrum sätt "Xpos": -50.0
