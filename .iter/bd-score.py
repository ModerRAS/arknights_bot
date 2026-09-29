from PIL import Image
import numpy as np
a = np.asarray(Image.open(r'C:/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes/src/utils/media/testdata/visual/final/old/box-detail.jpg').convert('RGB'), dtype=np.int32)
b = np.asarray(Image.open(r'C:/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes/.iter/bd-sweep.png').convert('RGB'), dtype=np.int32)
d = np.abs(a - b).sum()
print(1 - d / (a.shape[0]*a.shape[1]*3*255))
