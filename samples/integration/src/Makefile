# Compiler and linker options
CXX = mpicxx
CXXFLAGS = -Wall -O3
MPI_LFLAGS =
LDFLAGS = 

# Source files and object files
SRCS = main.cpp
OBJS = $(SRCS:.cpp=.o)

# Libraries and header file paths
LIBS = 
INCLUDES = 

# Target executable
EXEC = mpi-func

# Phony targets
.PHONY: all clean

# Build rules
all: $(EXEC)

$(EXEC): $(OBJS)
	$(CXX) $(MPI_LFLAGS) $(OBJS) $(LIBS) -o $(EXEC) $(LDFLAGS)

%.o: %.cpp
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

clean:
	rm -f $(OBJS) $(EXEC)