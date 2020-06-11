import os

num_tests = 500


if __name__ == "__main__":
    for i in range(50):
        # file_name = "out" + str(i)
        # os.system("go test -race -run 2C >" + file_name)
        # with open(file_name) as f:
        #     if 'FAIL' in f.read():
        #         print(file_name + " fails")
        #         continue
        #     else:
        #         print(file_name + " ok")

        # os.system("rm " + file_name)

        print("ROUND " + str(i))
        os.system("go test -race -run 2C")
