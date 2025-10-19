use crate::Message;

#[test]
fn test_tests() {
  assert!(true);
}

fn vecs_equal(vec1: Vec<String>, vec2: Vec<String>) -> bool {
  if vec1.len() != vec2.len() {
    return false;
  }

  let mut i = 0;
  while i < vec1.len() {
    if vec1[i] != vec2[i] {
      return false;
    }
    i += 1;
  }

  true
}

#[test]
fn test_split_into_lines() {
  let width = 5;

  let mut message = Message::default();
  message.body = "this is myy message".to_string();

  let output = message.split_into_lines(width);

  for line in &output {
    println!("{}", line);
  }

  let mut expected: Vec<String> = Vec::new();
  for line in vec!["this", "is", "myy", "messa", "ge"] {
    expected.push(line.to_string());
  }

  assert!(vecs_equal(output, expected))
}

#[test]
fn i_wanna_see() {
  let mut message = Message::default();
  message.body = "first message lets make this message super loooong jfjlkdsjafkldjaflk it was not long enough last time time to yap fr".to_string();
  let width = 76;

  let output = message.split_into_lines(width);

  for line in &output {
    println!("{}", line);
  }

  // assert!(false);
  assert!(true);
}
