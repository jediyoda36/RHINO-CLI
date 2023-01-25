FROM zhuhe0321/mpich:v2.0 as builder

ARG func_name ${func_name}
ARG file ${file}
ARG compile ${compile}

COPY ${file} .
RUN ${compile} ${file} -o ${func_name}


FROM zhuhe0321/mpirun_base

ARG func_name ${func_name}
COPY --from=builder /app/${func_name}  /app/${func_name}


CMD ["/bin/ash"]