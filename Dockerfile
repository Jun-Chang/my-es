FROM elasticsearch:8.12.1
COPY analysis/synonym.txt /usr/share/elasticsearch/config/analysis/synonym.txt
RUN elasticsearch-plugin install analysis-kuromoji
