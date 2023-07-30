import psycopg2
from psycopg2 import OperationalError


class Storage:
    """Storage class to work with the storage layer"""

    conn = None

    def __init__(self, host, port, user, password, name):
        connection = None

        try:
            connection = psycopg2.connect(
                database=name,
                user=user,
                password=password,
                host=host,
                port=port,
            )

            print("connectected to postgres")
        except OperationalError as e:
            print(f"connection to db error: {e}")
            exit(1)  # TODO: add correct error handling

        connection.autocommit = True

        self.conn = connection

        self.migrate()

    def migrate(self):
        """
        Check and update database schema to the current implementation
        TODO: use migration tool
        :return:
        """

        stmt = """
            create table if not exists word_count (
                user_id int,
                chat_id int,
                date date,
                val int,
                unique (user_id, chat_id, date)
            )
        """
        self.exec(stmt)

    def exec(self, query):
        """
        Execute given query on the storage database
        :param: query the query string
        :return: nothing
        """
        cursor = self.conn.cursor()
        try:
            cursor.execute(query)
        except OperationalError as e:
            print(f"error on query execution: {e}")
