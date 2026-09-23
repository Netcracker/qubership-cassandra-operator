This section provides information about Service Ports and Dependencies.

# List of Service Ports and Dependencies

Netcracker Cassandra Service consists Cassandra, Cassandra Backup Dameon, Cassandra Dbaas Adapter.
The list of service ports and dependencies are described in the following table:

|Component|Exposed Ports|Dependencies (Used Ports)|
|---------|-------------|-------------------------|
|Cassandra|`9042/TCP` The Cassandra server port, used by client applications to access database.|Cassandra Backup Daemon - `8080/TCP, 8443/TCP (TLS)`|
|Cassandra Backup Daemon|`8080/TCP, 8443/TCP (TLS)` The port for backup, restore, and obtaining status of database backups.|Cassandra - `9042/TCP`|
|Cassandra DBaaS Adapter|`8080/TCP, 8443/TCP (TLS)` The DBaaS API port, used by the DBaaS aggregator to manage cassandra database.|Cassandra - `9042/TCP`, Cassandra Backup Daemon - `8080/TCP, 8443/TCP (TLS)`, DBaaS Aggregator - `8080/TCP, 8443/TCP (TLS)`|
|Cassandra Prometheus Exporter|`9500/TCP`|Cassandra - `9042/TCP`, Cassandra Backup Daemon - `8080/TCP, 8443/TCP (TLS)`|


The following external interfaces are required for the service:

* DBaaS Aggregator 8080/TCP for registering physical database cluster.

# Cassandra Users Account deatils

## Admin

To connect to a Cassandra cluster using the cqlsh command-line tool, we can specify a username and password with the -u and -p options. This is particularly useful when authentication is enabled in Cassandra setup.

    cqlsh -u admin -p admin


# Cassandra Password Policy

Enabling Password Authentication
To enable password authentication in Cassandra, We need to configure the cassandra.yaml file


### Enable Authentication:

  `authenticator: PasswordAuthenticator`


### Enable Authorization (optional but recommended for more control)

  `authorizer: CassandraAuthorizer`


# Creating and Managing Users

### Create a Superuser:

  `CREATE ROLE admin WITH PASSWORD = 'your_secure_password' AND SUPERUSER = true AND LOGIN = true;`


### Create a Regular User:

  `CREATE ROLE username WITH PASSWORD = 'user_password' AND LOGIN = true;`


### Change a User's Password:

  `ALTER ROLE username WITH PASSWORD = 'new_password';`


### Drop a User:

  `DROP ROLE username;`



Cassandra doesn't enforce specific password policies by default. However, we can implement own policies through client applications or by integrating with external authentication systems.

For more detailed and specific configurations, refer to the official Apache Cassandra documentation.