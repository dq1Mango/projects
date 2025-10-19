use crate::Message;
use crate::Settings;

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
  let settings = &Settings {
    borders: true,
    message_width_ratio: 1.0,
  };

  let width = 5;

  let mut message = Message::default();
  message.body = "this is myy message".to_string();

  message.split_into_lines(settings, width);

  let raw_output = message.lines.unwrap();
  for line in &raw_output {
    println!("{}", line);
  }

  let mut expected: Vec<String> = Vec::new();
  for line in vec!["this", "is", "myy", "messa", "ge"] {
    expected.push(line.to_string());
  }
  assert!(vecs_equal(raw_output, expected))
}
