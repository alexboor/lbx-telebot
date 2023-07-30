# from datetime import date
import datetime
import utils
import nltk


def word_counter(storage, message):
    """
    Count word in given message object and store it to the storage
    :param storage: storage instance
    :param message: Message object (see telebot for more details)
    :return: None
    """
    uid = message.from_user.id
    cid = message.chat.id
    date = datetime.date.fromtimestamp(message.date)

    tokens = nltk.word_tokenize(utils.clean_text(message.text))
    tokens = [w for w in tokens if w not in nltk.corpus.stopwords.words("russian")]
    tokens = [w for w in tokens if w not in nltk.corpus.stopwords.words("english")]

    storage.count(uid, cid, date, len(tokens))
    return




