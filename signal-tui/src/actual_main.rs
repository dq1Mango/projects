use core::panic;
use std::{cmp::min, io::LineWriter, time::Duration};

use ratatui::{
  Frame,
  buffer::Buffer,
  crossterm::event::{self, Event, KeyCode},
  layout::Rect,
  style::Stylize,
  symbols::border,
  text::{Line, Text},
  widgets::{Block, Paragraph, Widget, Wrap},
};

#[derive(Debug, Default)]
struct Model {
  counter: i32,
  running_state: RunningState,
  content: Chat,
}

#[derive(Debug, Default, PartialEq, Eq)]
enum RunningState {
  #[default]
  Running,
  Done,
}

#[derive(PartialEq)]
enum Action {
  Increment,
  Decrement,
  Reset,
  Quit,
}

pub struct Settings {
  borders: bool,
  message_width_ratio: f32,
}

#[derive(Debug, Default, Clone)]
pub struct Message {
  body: String,
  lines: Option<Vec<String>>,
  sender: String,
}

#[derive(Default, Debug)]
pub struct Location {
  index: u64,
  offset: i16,
}

#[derive(Debug, Default)]
pub struct Chat {
  id: u8,
  messages: Vec<Message>,
  location: Location,
}

impl Model {
  fn init() -> Model {
    let messages = vec![
      Message {
        body: String::from("first message"),
        sender: String::from(""),
        lines: None,
      },
      Message {
        body: String::from("second message"),
        sender: String::from(""),
        lines: None,
      },
    ];

    let mut chat = Chat::default();

    for message in messages {
      chat.messages.push(message);
    }

    let mut app = Model::default();
    app.content = chat;
    app
  }
}

impl Message {
  fn render(self, area: Rect, buf: &mut Buffer, settings: Settings) {
    let block = Block::bordered().border_set(border::THICK);

    let availible_width: u16 = (area.width as f32 * 0.9).round() as u16;
    let mut new_area = area.clone();
    new_area.width = availible_width;
    //
    // let mut lines: Vec<String> = Vec::new();
    //
    // let mut index = 0;

    let result = self.lines;
    let lines: Vec<String>;

    match result {
      Some(x) => lines = x,
      None => {
        self.split_into_lines(settings, area.width);
        let this_better_work = self.lines;
        match this_better_work {
          Some(x) => lines = x,
          None => panic!("AAAaaaHHHhhh!!!"),
        }
      }
    }
    Paragraph::new(lines)
      .wrap(Wrap { trim: true })
      .block(block)
      .render(new_area, buf)
  }
}

// if you need more than 255 lines sue me
impl Message {
  fn split_into_lines(mut self, settings: Settings, term_width: u16) {
    let mut lines: Vec<String> = Vec::new();
    let mut new_line = String::from("");

    // collumn index
    let mut coldex = 0;
    let availible_width = (term_width as f32 * settings.message_width_ratio + 0.5) as usize;

    for yap in self.body.split(" ") {
      let mut length = yap.len();

      if coldex + yap.len() < availible_width {
        new_line.push_str(yap);
        new_line.push_str(" ");
        coldex += yap.len();
      } else {
        lines.push(new_line.clone().trim_end().to_string());
        coldex = 0;

        let mut index = 0;

        while index + length >= availible_width {
          lines.push(yap[index..index + availible_width].to_string());
          length -= availible_width;
          index += availible_width;
        }

        new_line = String::from("");
        new_line.push_str(" ");
      }
    }

    lines.push(new_line.clone().trim_end().to_string());

    self.lines = Some(lines);
  }
}

impl Widget for &Chat {
  fn render(self, area: Rect, buf: &mut Buffer) {
    let mut index = 0;
    let mut y = self.location.offset * -1;

    while y < area.height as i16 {
      let message = &self.messages[index];
      let height = message.lines.as_ref().unwrap().len() as i16;

      // let height = min(y + requested_height, area.height);
      let new_area = Rect::new(area.x, area.y - y as u16, area.width, height as u16);

      message.render(new_area, buf);

      index += 1;
      y += height;
    }
  }
}

fn main() -> color_eyre::Result<()> {
  // tui::install_panic_hook();
  let mut terminal = ratatui::init();
  let mut model = Model::init();

  while model.running_state != RunningState::Done {
    // Render the current view
    terminal.draw(|f| view(&mut model, f))?;

    // Handle events and map to a Message
    let mut current_msg = handle_event(&model)?;

    // Process updates as long as they return a non-None message
    while current_msg.is_some() {
      current_msg = update(&mut model, current_msg.unwrap());
    }
  }

  ratatui::restore();
  Ok(())
}

// TODO: gotta figure out how to model chat state

fn view(model: &mut Model, frame: &mut Frame) {
  let title = Line::from(" Counter App Tutorial ".bold());
  let instructions = Line::from(vec![
    " Decrement ".into(),
    "<Left>".blue().bold(),
    " Increment ".into(),
    "<Right>".blue().bold(),
    " Quit ".into(),
    "<Q> ".blue().bold(),
  ]);
  let block = Block::bordered()
    .title(title.centered())
    .title_bottom(instructions.centered())
    .border_set(border::THICK);

  let _counter_text = Text::from(vec![Line::from(vec![
    "Value: ".into(),
    model.counter.to_string().yellow(),
  ])]);

  let mut messages: Vec<ratatui::prelude::Line> = Vec::new();

  for message in model.content.messages.clone() {
    messages.push(Line::from(message.body));
  }

  let message_text = Text::from(messages);

  frame.render_widget(
    Paragraph::new(message_text).right_aligned().block(block),
    frame.area(),
  );

  // let p = Paragraph::new("A very long text that might not fit the container...")
  //   .wrap(Wrap { trim: true });
  //
  // let test_rect = Rect::new(10, 10, 7, 20);
  // frame.render_widget(p, test_rect);
}

/// Convert Event to Action
///
/// We don't need to pass in a `model` to this function in this example
/// but you might need it as your project evolves
fn handle_event(_: &Model) -> color_eyre::Result<Option<Action>> {
  if event::poll(Duration::from_millis(250))? {
    if let Event::Key(key) = event::read()? {
      if key.kind == event::KeyEventKind::Press {
        return Ok(handle_key(key));
      }
    }
  }
  Ok(None)
}

fn handle_key(key: event::KeyEvent) -> Option<Action> {
  match key.code {
    KeyCode::Char('j') => Some(Action::Increment),
    KeyCode::Char('k') => Some(Action::Decrement),
    KeyCode::Char('q') => Some(Action::Quit),
    _ => None,
  }
}

fn update(model: &mut Model, msg: Action) -> Option<Action> {
  match msg {
    Action::Increment => {
      model.counter += 1;
      if model.counter > 50 {
        return Some(Action::Reset);
      }
    }
    Action::Decrement => {
      model.counter -= 1;
      if model.counter < -50 {
        return Some(Action::Reset);
      }
    }
    Action::Reset => model.counter = 0,
    Action::Quit => {
      // You can handle cleanup and exit here
      model.running_state = RunningState::Done;
    }
  };
  None
}

#[test]
fn test() {
  println!("this is a test");
}
