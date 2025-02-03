FROM scratch

ARG TARGETPLATFORM

COPY /build/$TARGETPLATFORM/ /

ENTRYPOINT ["/airflow-prometheus-exporter"]
