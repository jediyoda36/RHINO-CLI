/* System and user includes */
#include <mpi.h>
#include <iostream>

using namespace std;

// define MPI tag
#define DATA_PACKET 0
#define RECV_RESULT 1
#define CALC_FINISH 2

/* Error handling */
#define MPI_CHECK(call) if((call) != MPI_SUCCESS) { \
    error_exit("MPI failed when calling " #call); \
}

void error_exit(const char *error) {
    cerr << error << endl;
    MPI_Abort(MPI_COMM_WORLD, -1);
}

// function to be integrated
double f(double x);

// use trapezoidal rule to compute numerical integration
void area_cal(double *result, double start, double end, double step);

// generate data packets
void generate_data(double **data, int interval, double start, double end);

int main(int argc, char *argv[])
{   
    /* Initialize the MPI environment */
    MPI_CHECK(MPI_Init(&argc, &argv));

    /* Get the number of processes, current rank and hostname */
    int world_size, world_rank, name_len;
    char processor_name[MPI_MAX_PROCESSOR_NAME];

    MPI_CHECK(MPI_Comm_size(MPI_COMM_WORLD, &world_size));
    MPI_CHECK(MPI_Comm_rank(MPI_COMM_WORLD, &world_rank));
    MPI_CHECK(MPI_Get_processor_name(processor_name, &name_len));

	/*
	 * YOUR CODE HERE
	 */
    int interval, i;
    double start, end, step, t_start, t_end;
    double temp = 0, result = 0;
    double *data;
    MPI_Status status;
    
    if (world_size < 2) {
        error_exit("Function needs at least two processes");
    }

    if (argc < 3) {
        error_exit("Function needs 3 Parameters: start(double), end(double), interval(int)");
    }
    start = atof(argv[1]);
    end = atof(argv[2]);
    interval = atoi(argv[3]);

    if (world_rank == 0){
        t_start = MPI_Wtime();
        generate_data(&data, interval, start, end);
        int counter = 0;

        // distribute data packets to worker
        for(i = 1; i < world_size; i++, counter++)
            MPI_CHECK(MPI_Send(&(data[2 * counter]), 2, MPI_DOUBLE, i, DATA_PACKET, MPI_COMM_WORLD));

        // wait for results
        do{
            MPI_CHECK(MPI_Recv(&temp, 1, MPI_DOUBLE, MPI_ANY_SOURCE, RECV_RESULT, MPI_COMM_WORLD, &status));
            result += temp;

            // send another data packet
            MPI_CHECK(MPI_Send(&(data[2 * counter]), 2, MPI_DOUBLE, status.MPI_SOURCE, DATA_PACKET, MPI_COMM_WORLD));
            counter++;

        }
        while (counter < 100 * interval); // do until no data packets to send

        // wait for pending results
        for(i = 1; i < world_size; i++) {
            MPI_CHECK(MPI_Recv(&temp, 1, MPI_DOUBLE, MPI_ANY_SOURCE, RECV_RESULT, MPI_COMM_WORLD, &status));
            result += temp;
        }

        // send finish message
        for(i = 1; i < world_size; i++)
        MPI_CHECK(MPI_Send(NULL, 0, MPI_DOUBLE, i, CALC_FINISH, MPI_COMM_WORLD));
        
        t_end = MPI_Wtime();

        printf("Result:  %f\n", result);
        printf("Time: %.2fs\n", t_end - t_start);
    } 
    else {
        double data[2];
        do {
            MPI_CHECK(MPI_Probe(0, MPI_ANY_TAG, MPI_COMM_WORLD, &status));
            if(status.MPI_TAG == DATA_PACKET){      
                MPI_CHECK(MPI_Recv(data, 2, MPI_DOUBLE, 0, DATA_PACKET, MPI_COMM_WORLD, &status));
                step = (data[1] - data[0]) / 40960000;
                area_cal(&temp, data[0], data[1], step);
                MPI_CHECK(MPI_Send(&temp, 1, MPI_DOUBLE, 0, RECV_RESULT, MPI_COMM_WORLD));
	        }
        } while (status.MPI_TAG != CALC_FINISH);
    }

    MPI_CHECK(MPI_Finalize());
    return EXIT_SUCCESS;
}

double f(double x){
    return 1.0 / (1.0 + x);
}

void area_cal(double *result, double start, double end, double step){
    double area = 0;
    double begin = start;
    long count;
    long count_max = ((end - start) / step);

    for(count = 0; count < count_max; count++){
      area += step * (f(begin) + f(begin + step)) / 2;
      begin += step;
    }     
    *result = area;
}

void generate_data(double **data, int interval, double start, double end){
    long i;
    int packet_count = 100 * interval;
    double num = start;
    double step = (end - start) / (double)(packet_count);
    *data = (double *)malloc(sizeof(double) * packet_count * 2);
    if (!(*data))
        error_exit("Not enough memory when generating data");
    for(i = 0; i < 2 * packet_count; ){
        (*data)[i++] = num;
        num += step;
        (*data)[i++] = num;
    }
}
