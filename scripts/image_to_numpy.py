import numpy as np
from PIL import Image
import sys

if len(sys.argv) > 1: 
    input = sys.argv[1]
else:
    input = ""

image = Image.open(input) 
image_array = np.array(image)

np.savetxt("image_data.csv", image_array.reshape(-1, image_array.shape[-1]), delimiter=",", fmt="%d")
# np.save("image_data.npy", image_array)