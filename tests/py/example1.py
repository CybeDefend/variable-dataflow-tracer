import os

# Fonction DataFlowTest
def DataFlowTest(filePath, test):
    newPath = "example backward"

    # Vérifie si le fichier existe
    if not os.path.exists(newPath):
        return "File does not exist"
    else:
        # Lis le contenu du fichier
        try:
            with open(newPath, 'r') as file:
                return file.read()
        except:
            return "Error reading file"
        
    filePathModified = "1" + newPath

    filePathModified = "test"
    
    return result

# Fonction functionTest
def functionTest():
    return "example backward"

# Fonction TEST2
def TEST2(test):
    test = "example testAAA"
    return test

# Fonction test
def test():
    filePath = "example.txt"
    if filePath == "":
        print("File does not exist")

    testStr = "test"
    TEST2(filePathModified)

    filePathModified = filePathModified + "1"
    test = "test"
    message = DataFlowTest(filePathModified, test)

    print(message)

# Fonction principale
def main():
    filePath = "example backward"
    test()

main()
