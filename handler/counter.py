# from datetime import date
import datetime
import utils
import nltk


def word_counter(message):
    """
    Count word in given message object and store it to the storage
    :param message: Message object (see telebot for more details)
    :return: None
    """
    user_id = message.from_user.id
    chat_id = message.chat.id
    date = datetime.date.fromtimestamp(message.date)

    tokens = nltk.word_tokenize(utils.clean_text(message.text))
    tokens = [w for w in tokens if w not in nltk.corpus.stopwords.words("russian")]
    tokens = [w for w in tokens if w not in nltk.corpus.stopwords.words("english")]

    num = len(tokens)

    print(f"user {user_id} in chat {chat_id} at {date}: count = {num}")
    print(tokens)




