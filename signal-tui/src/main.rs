mod logger;
mod multi_line_string;
#[cfg(test)]
mod tests;

mod update;

use std::{time::Duration, vec};

use ratatui::{
  Frame,
  buffer::Buffer,
  layout::{Constraint, Direction, Layout, Position, Rect},
  style::Stylize,
  symbols::border,
  text::Line,
  widgets::{Block, Paragraph, Widget},
};

use crate::logger::Logger;
use crate::multi_line_string::MultiLineString;
use crate::update::*;

#[derive(Debug, Default)]
pub struct Model {
  running_state: RunningState,
  mode: Mode,
  chats: Vec<Chat>,
  chat_index: usize,
}

#[derive(Debug, Default, PartialEq, Eq)]
pub enum RunningState {
  #[default]
  Running,
  OhShit,
}

#[derive(Default, Debug, PartialEq)]
pub enum Mode {
  #[default]
  Normal,
  Insert,
}

#[derive(Debug, Default, Clone)]
pub struct Message {
  body: MultiLineString,
  sender: String,
}

#[derive(Default, Debug)]
pub struct Location {
  index: u64,
  offset: i16,
}

#[derive(Debug, Default)]
pub struct Chat {
  // id: u8,
  messages: Vec<Message>,
  location: Location,
  text_input: TextInput,
}

#[derive(Debug, Default)]
pub struct TextInput {
  body: MultiLineString,
  cursor_index: u16,
  cursor_position: Position,
}

pub struct Settings {
  borders: bool,
  message_width_ratio: f32,
  identity: String,
}

impl Settings {
  fn init() -> Self {
    Self {
      borders: true,
      message_width_ratio: 0.8,
      identity: "me".to_string(),
    }
  }
}

impl Model {
  fn init() -> Self {
    let messages = vec![
      Message {
        body: MultiLineString::init(
          "first message lets make this message super looong jjafkldjaflk it was not long enough last time time to yap fr",
        ),
        sender: String::from("not me"),
      },
      Message {
        body: MultiLineString::init("second message"),
        sender: String::from("me"),
      },
    ];

    let mut chat = Chat::default();

    for message in messages {
      chat.messages.push(message);
    }

    chat.text_input = TextInput::default();

    let chats: Vec<Chat> = vec![chat];

    let mut model = Model::default();
    model.chats = chats;
    model.chat_index = 0;
    model
  }

  // not really needed but it staves off the need for explicit liiftimes a little longer
  fn current_chat(&mut self) -> &mut Chat {
    &mut self.chats[self.chat_index]
  }
}

impl TextInput {
  fn render(&mut self, area: Rect, buf: &mut Buffer, _logger: &mut Logger) {
    let block = Block::bordered().border_set(border::THICK);

    // shitty temp padding for the border
    // let mut area = area;
    // area.x += 1;
    // area.width -= 2;
    // area.height -= 2;
    // area.y += 1;

    // minus 3 b/c you cant have the cursor on the border and i cant be bothered to add another
    // edge case
    let vec_lines = self.body.as_lines(area.width - 3).to_vec();
    // logger.log(format!("this is the first line: {}", self.cursor_index));
    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(area, buf);

    self.cursor_position = self.calc_cursor_position(area)
  }

  fn calc_cursor_position(&mut self, area: Rect) -> Position {
    // gotta pad the border (still havent found a better way of doing this)
    let mut pos = Position {
      x: area.x + 1,
      y: area.y + 1,
    };
    // mad ugly calculations, smthns gotta change
    let lines = self.body.as_lines(area.width - 3);
    // let body = self.body.body.char_indices();

    let (mut index, mut row, mut col) = (0, 0, 0);

    while index + lines[row].len() < self.cursor_index as usize {
      index += lines[row].len();
      pos.y += 1;
      row += 1;
    }

    pos.x += self.cursor_index - index as u16;

    // let length = self.body.body.len() as u16;

    let mut y = area.y + 1;
    // if lines != 0 {
    //   y += self.cursor_index / (lines + 1) as u16
    // }

    let x = area.x + 1 + self.cursor_index / area.width;

    pos
  }

  fn insert_char(&mut self, char: char, _logger: &mut Logger) {
    // some disgusting object-oriented blashphemy going on here
    self.body.body.insert(self.cursor_index as usize, char);
    self.cursor_index += 1;
  }

  fn delete_char(&mut self) {
    if self.cursor_index == 0 {
      return;
    }

    self.cursor_index -= 1;
    self.body.body.remove(self.cursor_index as usize);
  }
}

impl Message {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, _logger: &mut Logger) {
    let block = Block::bordered().border_set(border::THICK);

    // this ugly shadow cost me a good 15 mins of my life ... but im not changing it
    let mut my_area = area.clone();
    my_area.width = (area.width as f32 * settings.message_width_ratio + 0.5) as u16;
    // let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;

    // let mut new_area = area.clone();
    // new_area.width = area.width;
    //
    // let mut lines: Vec<String> = Vec::new();
    //
    // let mut index = 0;

    // let result = &self.lines;

    let vec_lines: Vec<String> = self.body.as_lines(my_area.width - 2).to_vec();

    // match result {
    //   Some(x) => vec_lines = x.to_vec(),
    //   None => {
    //     panic!("AAAaaaHHHhhh!!!")
    //   }
    // }

    // shrink the message to fit if it does not need mutliple lines
    if vec_lines.len() == 1 {
      my_area.width = vec_lines[0].len() as u16 + 2;
    }

    // "allign" the chat to the right if it was sent by you
    // TODO: should add setting to toggle this behavior
    if settings.identity == self.sender {
      my_area.x += area.width - my_area.width;
    }

    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(my_area, buf)
    // .wrap(Wrap { trim: true })
  }
}

fn format_vec(vec: &Vec<String>) -> String {
  let mut output = String::from("[");

  for thing in vec {
    output.push_str(thing);
    output.push_str(", ");
  }

  output.push_str("]");

  return output;
}

impl Chat {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, logger: &mut Logger) {
    let input_lines = self.text_input.body.as_lines(area.width - 2).len() as u16;
    Logger::log("this is our input: ".to_string());
    Logger::log(format_vec(self.text_input.body.as_lines(area.width - 2)));

    let layout = Layout::vertical([Constraint::Min(6), Constraint::Length(input_lines + 2)]).split(area);

    // kind of a sketchy shadow here but the layout[1] is used like once
    let area = layout[0];

    let block = Block::bordered().border_set(border::THICK);
    // .title(title.centered())
    // .title_bottom(instructions.centered())
    block.render(area, buf);

    // shitty temp padding for the border
    let mut area = area;
    area.x += 1;
    area.width -= 2;
    area.height -= 2;
    area.y += 1;
    // end shitty tmp padding

    let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;

    let mut index = 0;
    let mut y = self.location.offset * -1;

    while y < area.height as i16 && index < self.messages.len() {
      let message = &mut self.messages[index];

      // only here for testing; remove later; breaks the fun "cache" system
      // message.lines = message.as_lines(message_width);

      // let result = message.body.as_lines(mes);

      // match result {
      //   Some(_x) => {}
      //   None => {
      //     message.lines = Some(message.split_into_lines(message_width));
      //     let this_better_work = &message.lines;
      //     match this_better_work {
      //       Some(_x) => {}
      //       None => panic!("AAAaaaHHHhhh!!!"),
      //     }
      //   }
      // }

      let height = message.body.as_lines(message_width).len() as i16 + 2;

      // let height = min(y + requested_height, area.height);
      let new_area = Rect::new(area.x, area.y + y as u16, area.width, height as u16);

      message.render(new_area, buf, settings, logger);

      index += 1;
      y += height;
    }

    self.text_input.render(layout[1], buf, logger);
  }
}

fn main() -> color_eyre::Result<()> {
  // tui::install_panic_hook();
  let mut terminal = ratatui::init();
  let mut model = Model::init();
  let settings = &Settings::init();

  // regular lumber jack
  let logger = &mut Logger::init("log.txt");
  Logger::log("testing".to_string());

  while model.running_state != RunningState::OhShit {
    // Render the current view
    terminal.draw(|f| view(&mut model, f, settings, logger))?;

    // Handle events and map to a Message
    let mut current_msg = handle_event(&model)?;

    // Process updates as long as they return a non-None message
    while current_msg.is_some() {
      current_msg = update(&mut model, current_msg.unwrap(), logger);
    }
  }

  ratatui::restore();
  Ok(())
}

// TODO: gotta figure out how to model chat state

fn view(model: &mut Model, frame: &mut Frame, settings: &Settings, logger: &mut Logger) {
  let title = Line::from(" Counter App Tutorial ".bold());
  let instructions = Line::from(vec![
    " Decrement ".into(),
    "<Left>".blue().bold(),
    " Increment ".into(),
    "<Right>".blue().bold(),
    " Quit ".into(),
    "<Q> ".blue().bold(),
  ]);
  let _block = Block::bordered()
    .title(title.centered())
    .title_bottom(instructions.centered())
    .border_set(border::THICK);

  // let _counter_text = Text::from(vec![Line::from(vec![
  //   "Value: ".into(),
  //   model.counter.to_string().yellow(),
  // ])]);

  let layout = Layout::new(
    Direction::Horizontal,
    vec![Constraint::Percentage(20), Constraint::Percentage(80)],
  )
  .split(frame.area());

  model
    .current_chat()
    .render(layout[1], frame.buffer_mut(), settings, logger);

  frame.set_cursor_position(model.current_chat().text_input.cursor_position);

  //
  // frame.render_widget(
  //   Paragraph::new(message_text).right_aligned().block(block),
  //   frame.area(),
  // );

  // let p = Paragraph::new("A very long text that might not fit the container...")
  //   .wrap(Wrap { trim: true });
  //
  // let test_rect = Rect::new(10, 10, 7, 20);
  // frame.render_widget(p, test_rect);
}
