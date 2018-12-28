# Användarinstruktion:
Editera JSON filen så den passar materialet.
Skapa fräsfiler. En nära den rörliga infästningen och en nära den fasta infästningen
Fräs båda filerna, vänd materialet och fräs samma filer på andra sidan
Mät tjockleken på det utfrästa materialet: 

För Z gäller: 
  Rätt kalibrering ger en tjocklek på 80 mm.
  Om tjockleken är 84 mm ska fläsningen flyttas ner 2 mm. Sätt då HomingOffset[2] till +2 mm.
För X gäller:
  Felet i x är halva skilladen mellan vertikala planen x och x-.
  Korrgera x med att ändra HomingOffset[0] +-felet. Plus eller minus beror på hur skillnaden ser ut.
  Om det fräsen har fräst för långt bort i x-led, ska minus användas. 
  Om fräsen har fräst för nära i x-led, ska plus användas.

Parametrar i json filen:
ToolRadius, radien på fräsverktyget
MaxHeight, det högsta tillåtena z värdet. Samma som HomingOffset[2]
Xpos, positionen i xled som avstånd från centrum
Ypos, positionen i yled 
MaterialThickness, tjockleken på materialet
FilePrefix, om man vill ge ett mer specifikt namn på g-kod filen

För att köra en fil som gör ett plan i centrum sätt "Xpos": -50.0
