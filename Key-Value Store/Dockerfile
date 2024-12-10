FROM python:3

#Here is the working directory for the dockerfile
WORKDIR /usr/src/app

#Copying main.py into dockerfile
COPY main.py ./

#Installing flask dependency
RUN pip install flask
RUN pip install requests

#Setting flask port
ENV FLASK_RUN_PORT=8090

#Command line to run
CMD [ "flask", "--app", "main", "run", "--host=0.0.0.0"]