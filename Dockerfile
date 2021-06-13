FROM ubuntu:latest

# web port
EXPOSE 8081
# grpc port
EXPOSE 50051

ADD build/authGo /app/authGo
RUN ls -l

CMD /app/authGo
