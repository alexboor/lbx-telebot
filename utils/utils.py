import string


def clean_text(text):
    """
    Clean given text for further processing
    It make words to lower case, remove punctuation and spec symbols and digits
    :param text: raw text
    :return: clean text
    """
    res = text.lower()
    res = remove(res, string.punctuation)
    res = remove(res, string.digits)

    return res


def remove(text, chars):
    """
    Removing chars from given text
    :param text:
    :param chars:
    :return:
    """
    return "".join([ch for ch in text if ch not in chars])

