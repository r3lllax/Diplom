install acl
<hr>

<b>./photos - any mounting directory</b>

<p>sudo setfacl -R -m u:101:rx ./photos</p>
<p>sudo setfacl -R -m u:101:r ./photos/*</p>

<p>sudo setfacl -R -m u:$(id -u):rwx ./photos</p>
<p>sudo setfacl -R -m u:$(id -u):rw ./photos/*</p>

<p>sudo setfacl -R -d -m u:101:rx ./photos</p>
<p>sudo setfacl -R -d -m u:$(id -u):rwx ./photos</p>