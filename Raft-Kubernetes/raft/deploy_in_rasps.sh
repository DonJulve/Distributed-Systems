ssh a840710@192.168.3.5 'rm -r /home/a840710/practica4'
ssh a840710@192.168.3.6 'rm -r /home/a840710/practica4'
ssh a840710@192.168.3.7 'rm -r /home/a840710/practica4'
ssh a840710@192.168.3.8 'rm -r /home/a840710/practica4'
scp -r /home/a840710/practica4 a840710@192.168.3.5:/home/a840710/
scp -r /home/a840710/practica4 a840710@192.168.3.6:/home/a840710/
scp -r /home/a840710/practica4 a840710@192.168.3.7:/home/a840710/
scp -r /home/a840710/practica4 a840710@192.168.3.8:/home/a840710/
