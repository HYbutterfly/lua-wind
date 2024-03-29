CFLAG = -Wall -g
INC = -I3rd/lua-5.4.6/src -Isrc -Iluaclib-src
LIB = -L3rd/lua-5.4.6/src -llua -lm -DLUA_USE_READLINE -ldl -pthread

SRC = main.c \
	queue.c \

LUALIB = lua_wind.c \
	lua_serialize.c \
	lua_socket.c \
	lua_epoll.c \
	lua_eventfd.c \
	lua_timerfd.c \


.PHONY: wind

all:
	$(CC) $(CFLAG) -o wind $(addprefix src/,$(SRC)) $(addprefix luaclib-src/,$(LUALIB)) $(INC) $(LIB)