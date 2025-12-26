import cv2
import os

def drawLine(name, img, detections, h, w):
    # Draw bounding boxes on faces
    for i in range(0, detections.shape[2]):
        confidence = detections[0, 0, i, 2]

        # Filter out weak detections
        if confidence > 0.5:
            box = detections[0, 0, i, 3:7] * [w, h, w, h]
            (x1, y1, x2, y2) = box.astype("int")
            cv2.rectangle(img, (x1, y1), (x2, y2), (0, 255, 0), 2)
            cv2.putText(img, f"{confidence:.2f}", (x1, y1-10),
                        cv2.FONT_HERSHEY_SIMPLEX, 0.5, (0, 255, 0), 2)

    # Save the output
    saved = f"./detected/{name}.jpg"
    cv2.imwrite(saved, img)
    print(f"Saved {name} with {detections.shape[2]} total detections")

# Paths to the model files
modelFile = "res10_300x300_ssd_iter_140000.caffemodel"
configFile = "deploy.prototxt"

# Load the network
net = cv2.dnn.readNetFromCaffe(configFile, modelFile)

for i in range(30):
    # Load the image
    i = i+1
    name = f"{i:06}"
    file = f"./train/{name}.jpg"
    if i <= 10:
        file = f"./train/{name}.png"

    img = cv2.imread(file)
    (h, w) = img.shape[:2]

    # Prepare the blob
    blob = cv2.dnn.blobFromImage(cv2.resize(img, (300, 300)), 1.0,
                                (300, 300), (104.0, 177.0, 123.0))

    # Forward pass to get detections
    net.setInput(blob)
    detections = net.forward()

    drawLine(name, img, detections, h, w)
    



