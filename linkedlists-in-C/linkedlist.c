#include "stdio.h"
#include "stdlib.h"

// TODO: refactor functions to use a crawling function to get the correct index;
#define len(x) (sizeof(x) / sizeof((x)[0]))

struct element {
  struct element *prev;
  int data;
  struct element *next;
};

int indexOutOfRangeError() {
  printf("Index out of range error");
  return -1;
}

struct element *initialize(int initial[], int length) {

  struct element *prev;
  struct element *patientZero;

  patientZero = malloc(sizeof(struct element));
  patientZero->data = initial[0];
  prev = patientZero;

  // printf("%i", len(initial));

  for (int i = 1; i < length; i++) {

    struct element *newItem;

    newItem = malloc(sizeof(struct element));
    newItem->data = initial[i];
    newItem->prev = prev;

    prev->next = newItem;
    prev = newItem;
  }

  return patientZero;
};

int read(struct element *list, int index) {

  if (index < 0) {
    return indexOutOfRangeError();
  }

  struct element *node = list;

  for (int i = 0; i < index; i++) {
    if (node->next == NULL) {
      return indexOutOfRangeError();
    }

    node = node->next;
  }

  return node->data;
}

int write(struct element *list, int data, int index) {

  if (index < 0) {
    return indexOutOfRangeError();
  }

  struct element *node = list;

  for (int i = 0; i < index; i++) {
    if (node->next == NULL) {
      return indexOutOfRangeError();
    }

    node = node->next;
  }

  node->data = data;
  return 0;
}

int length(struct element *list) {

  int length = 1;
  struct element *node = list;

  while (node->next != NULL) {
    node = node->next;
    length++;
  }

  return length;
}

int insert(struct element *list, int data, int index) {

  if (index < 0) {
    return indexOutOfRangeError();
  };

  // crawl the list to the correct index
  struct element *node = list;

  // initialize new node with data
  struct element *newItem;
  newItem = malloc(sizeof(struct element));

  if (index != 0) {
    newItem->data = data;
  } else {
    // edge case of inserting to the start of the list
    newItem->data = node->data;
    node->data = data;
    newItem->prev = node;
    newItem->next = node->next;
    node->next->prev = newItem;
    node->next = newItem;
    return 0;
  }

  for (int i = 0; i < index; i++) {

    // edge case of appending to the end of the list
    if (node->next == NULL) {
      if (index == i + 1) {

        newItem->prev = node;
        node->next = newItem;

        return 0;

      } else {
        printf("Index out of range for insert\n");
        return 1;
      }
    }

    node = node->next;
  }

  // typical operation (insertion in the middle of the list)
  newItem->next = node;
  newItem->prev = node->prev;

  newItem->prev->next = newItem;
  node->prev = newItem;

  return 0;
}

int delete(struct element *list, int index) {

  if (index < 0) {
    return indexOutOfRangeError();
  }

  struct element *node = list;

  // deletion of first item
  if (index == 0) {
    node->data = node->next->data;
    index = 1;
  }
  // if (index == 0) {
  //  node = list->next;
  //  list->data = node->data;
  //  list->next = data->next
  //}

  for (int i = 0; i < index; i++) {
    if (node->next == NULL) {
      return indexOutOfRangeError();
    }

    node = node->next;
  }

  if (node->next != NULL) {
    // deletion of middle item
    node->prev->next = node->next;
    node->next->prev = node->prev;
  } else {
    // deletion of last item
    node->prev->next = NULL;
  }

  free(node);

  return 0;
}

int *copy(struct element *list) {

  int *array;
  int len = length(list);
  array = calloc(sizeof(int), len);

  struct element *node = list;
  array[0] = node->data;

  for (int i = 1; i < len; i++) {
    node = node->next;
    array[i] = node->data;
  }

  return array;
}

int clear(struct element *list) {
  struct element *node = list;
  struct element *next;
  while (node->next != NULL) {
    next = node->next;
    free(node);
    node = next;
  }

  free(node);
  return 0;
}

void print(struct element *list) {

  struct element current = *list;

  printf("%s", "[");
  printf("%i", current.data);

  // feels like there is a cleaner way to do this
  while (current.next != NULL) {
    printf("%s", ", ");
    current = *current.next;
    printf("%i", current.data);
  }

  printf("%s", "]\n");
}

int main(int argc, char *argv[]) {

  int boringList[] = {0, 1, 2, 3, 4, 5, 6, 7, 67};

  struct element *funList = initialize(boringList, len(boringList));

  print(funList);
  insert(funList, 69, length(funList));
  print(funList);
  delete (funList, 9);
  print(funList);
  write(funList, 67, 1);
  print(funList);
  printf("%i", read(funList, 1));
  printf("\n");
  int *copied;
  copied = copy(funList);
  for (int i = 0; i < length(funList); i++) {
    printf("%d", copied[i]);
    printf(",");
  }
  puts("");
  printf("%p\n", (void *)&funList);
  // printf("%i\n", funList);
  clear(funList);

  printf("%p\n", (void *)&funList);
  // printf("%i", funList);
  return 0;
}
