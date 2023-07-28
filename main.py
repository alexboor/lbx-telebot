import os

import telebot

API_TOKEN = os.environ.get("TELEGRAM_BOT_TOKEN")

bot = telebot.TeleBot(API_TOKEN)

# Handle all other messages with content_type 'text' (content_types defaults to ['text'])
@bot.message_handler(func=lambda message: True)
def echo_message(message):
    bot.reply_to(message, message.text)


bot.infinity_polling()