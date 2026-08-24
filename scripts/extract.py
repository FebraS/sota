import sys
import os
from pyzbar.pyzbar import decode
from PIL import Image

def getOtpUri(image_path):
    try:
        if not os.path.isfile(image_path):
            print(f"Error: File '{image_path}' not found.", file=sys.stderr)
            return None

        img = Image.open(image_path)
        decodedObjects = decode(img)

        for obj in decodedObjects:
            uri = obj.data.decode('utf-8')
            if uri.startswith("otpauth-migration://") or uri.startswith("otpauth://"):
                return uri

        return None
    except Exception as e:
        print(f"Error processing image: {e}", file=sys.stderr)
        return None

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python scripts/extract.py <image_path>", file=sys.stderr)
        sys.exit(1)
    
    result = getOtpUri(sys.argv[1])
    if result:
        print(result)
    else:
        sys.exit(1)