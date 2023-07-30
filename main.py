import os
import telebot
from storage import Storage

API_TOKEN = os.environ.get("TELEGRAM_BOT_TOKEN")
VERSION = "1.0.3"

storage = Storage(
    os.environ.get("DB_HOST"),
    os.environ.get("DB_POSR"),
    os.environ.get("DB_USER"),
    os.environ.get("DB_PASSWORD"),
    os.environ.get("DB_NAME"),
)


bot = telebot.TeleBot(API_TOKEN)

# @bot.message_handler(commands=['ver', 'v'])
# def send_version(msg):
#     bot.reply_to(msg, VERSION)
#
# # Handle all other messages with content_type 'text' (content_types defaults to ['text'])
# @bot.message_handler(func=lambda message: True)
# def echo_message(message):
#     print(message)
#     bot.reply_to(message, message.text)







bot.infinity_polling()