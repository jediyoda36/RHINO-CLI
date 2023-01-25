ARG user ${user}
ARG name ${name}
ARG version ${version}
FROM ${user}/${name}_base:${version} as builder

ARG name ${name}
ENV FUNC_NAME=${name}

COPY util/ldd.sh .
RUN sh ldd.sh

FROM zhuhe0321/mpirun_base

ARG name ${name}
COPY --from=builder /app/${name} /app/${name}
COPY --from=builder /shared_lib /usr/local/lib

CMD ["/bin/ash"]