FROM scratch

COPY ./airflow-prometheus-exporter /airflow-prometheus-exporter
ENTRYPOINT ["/airflow-prometheus-exporter"]
