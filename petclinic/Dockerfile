FROM amazoncorretto:17

WORKDIR /app

COPY target/*.jar petclinic.jar

EXPOSE 8080

CMD [ "java", "-jar", "petclinic.jar" ]
