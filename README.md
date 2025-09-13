## Integrantes:
- Jhossep Martinez / 202173530-5
- Fernando Xais / 202273551-1
- Gabriela Yáñez / 202273511-2

## Consideraciones:
- La maquina virtual de lester (dist013) tiene rabbitMQ corriendo por lo que no es necesario resetearlo
- En el directorio de Lester se encuentra el archivo "ofertas_grandes.cvs", desde aquí se leen las ofertas y se ofrecen constantemente de forma aleatoria cada vez que michael solicita una. En caso de querer usar otro archivo este debe ir en el directorio de michael, debe cambiarse el nombre del archivo a cargar en el main de lester y debe volver a compilarse la maquina de lester.

## Instrucciones:
- Ir a la VM dist13 y ejecutar ```make docker-run-lester```
- Ir a la VM dist16 y ejecutar ```make docker-run-franklin```
- Ir a la VM dist15 y ejecutar ```make docker-run-trevor```
- Ir a la VM dist14 y ejecutar ```make docker-run-michael```



Credenciales VM:
dist013
ehe6gqRsS2Fk
10.35.168.23
lester
50051

dist016
jrKU59Umn2TW
10.35.168.26
Franklin
50054

dist015
aNASDGkYnQ8F
10.35.168.25
Trevor
50053

dist014
KRZ65kfAEmpB
10.35.168.24
Michael
50052
