FROM ubuntu:latest

# Essential for using tls
RUN apt-get update
RUN apt-get install ca-certificates -y
RUN update-ca-certificates

ENV AZURE-KEYVAULT-NAME=kotlang-secrets

# web port
EXPOSE 8081
# grpc port
EXPOSE 50051

ADD build/authGo /app/authGo
RUN ls -l

CMD /app/authGo
