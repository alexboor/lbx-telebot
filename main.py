import os
import telebot
import handler
from storage import Storage

API_TOKEN = os.environ.get("TELEGRAM_BOT_TOKEN")
VERSION = "1.1.5"
ALLOW_CHATS = [int(i) for i in os.environ.get("ALLOW_CHATS").split(",")]

print(f"Allowed chats: {ALLOW_CHATS}")

storage = Storage(
    os.environ.get("DB_HOST"),
    os.environ.get("DB_PORT"),
    os.environ.get("DB_USER"),
    os.environ.get("DB_PASSWORD"),
    os.environ.get("DB_NAME"),
)

bot = telebot.TeleBot(API_TOKEN)


@bot.message_handler(commands=['ver', 'v'])
def send_version(msg):
    bot.reply_to(msg, VERSION)


# Handle all other messages with content_type 'text' (content_types defaults to ['text'])
@bot.message_handler(func=lambda message: True)
def echo_message(message):
    if message.chat.id in ALLOW_CHATS:
        print(message)
        handler.word_counter(storage, message)


# def listener(messages):
#     for m in messages:
#         print(m.chat)
#
#
# bot.set_update_listener(listener)
bot.infinity_polling()
