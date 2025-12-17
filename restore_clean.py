#!/usr/bin/env python3
import psycopg2
import sys

conn_string = "postgresql://postgres:CgorBwkaFQBpvhIwtMVhHNXAXmPJZtMp@nozomi.proxy.rlwy.net:43580/railway"

try:
    conn = psycopg2.connect(conn_string)
    conn.autocommit = True
    cur = conn.cursor()
    
    print("🗑️  Clearing existing matches...")
    cur.execute("DELETE FROM matches;")
    
    print("🗑️  Clearing existing rounds...")
    cur.execute("DELETE FROM rounds;")
    
    print("📝 Creating rounds...")
    for i in range(1, 9):
        format_type = 'PB' if i % 2 == 1 else 'BF'
        cur.execute("INSERT INTO rounds (round_number, format) VALUES (%s, %s)", (i, format_type))
        print(f"  ✓ Round {i} ({format_type}) inserted")
    
    # Verify rounds were inserted
    cur.execute("SELECT COUNT(*) FROM rounds;")
    round_count = cur.fetchone()[0]
    print(f"✅ Verified: {round_count} rounds in database")
    
    # Get the round mapping
    cur.execute("SELECT id, round_number FROM rounds ORDER BY round_number;")
    rounds = cur.fetchall()
    round_map = {rnum: rid for rid, rnum in rounds}
    print(f"✅ Round mapping: {round_map}")
    
    print("🏆 Inserting matches...")
    matches_data = [
        # Round 1 (PB)
        (1, 150, 152, 2, 0), (1, 145, 147, 1, 2), (1, 148, 154, 1, 1), (1, 149, 151, 2, 0), (1, 153, 146, 2, 0),
        # Round 2 (BF)
        (2, 153, 152, 0, 2), (2, 149, 154, 1, 2), (2, 146, 151, 2, 0), (2, 150, 147, 1, 2), (2, 145, 148, 0, 2),
        # Round 3 (PB)
        (3, 153, 151, 2, 0), (3, 152, 147, 1, 2), (3, 146, 154, 1, 0), (3, 150, 148, 2, 0), (3, 149, 145, 2, 0),
        # Round 4 (BF)
        (4, 153, 147, 0, 2), (4, 152, 148, 2, 0), (4, 151, 154, 0, 2), (4, 150, 149, 2, 1), (4, 146, 145, 2, 1),
        # Round 5 (PB)
        (5, 153, 154, 1, 1), (5, 147, 148, 2, 0), (5, 151, 145, 0, 2), (5, 146, 150, 1, 1), (5, 152, 149, 0, 1),
        # Round 6 (BF)
        (6, 153, 148, 2, 0), (6, 154, 145, 2, 0), (6, 147, 149, 0, 2), (6, 151, 150, 0, 2), (6, 152, 146, 1, 1),
        # Round 7 (PB)
        (7, 153, 145, 1, 1), (7, 148, 149, 0, 2), (7, 154, 150, 1, 1), (7, 147, 146, 2, 1), (7, 151, 152, 0, 2),
        # Round 8 (BF)
        (8, 153, 149, 1, 1), (8, 145, 150, 2, 1), (8, 148, 146, 0, 2), (8, 154, 152, 0, 2), (8, 147, 151, 2, 0)
    ]
    
    for round_num, p1, p2, s1, s2 in matches_data:
        round_id = round_map[round_num]
        cur.execute(
            "INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES (%s, %s, %s, %s, %s, true)",
            (round_id, p1, p2, s1, s2)
        )
    
    print("✅ Matches inserted!")
    
    # Now populate player_match_stats for each match
    print("📊 Populating player_match_stats...")
    cur.execute("SELECT id, player1_id, score1, player2_id, score2 FROM matches WHERE completed = true;")
    matches = cur.fetchall()
    
    for match_id, p1_id, p1_score, p2_id, p2_score in matches:
        # Player 1 stats: games_played = sum of both scores, games_won = player1's score
        cur.execute(
            "INSERT INTO player_match_stats (player_id, match_id, games_played, games_won) VALUES (%s, %s, %s, %s)",
            (p1_id, match_id, p1_score + p2_score, p1_score)
        )
        # Player 2 stats: games_played = sum of both scores, games_won = player2's score
        cur.execute(
            "INSERT INTO player_match_stats (player_id, match_id, games_played, games_won) VALUES (%s, %s, %s, %s)",
            (p2_id, match_id, p1_score + p2_score, p2_score)
        )
    
    print("✅ Player match stats populated!")
    
    # Verify matches
    cur.execute("SELECT COUNT(*) FROM matches;")
    match_count = cur.fetchone()[0]
    print(f"✅ Verified: {match_count} matches in database")
    
    # Final verification - check rounds persist
    cur.execute("SELECT COUNT(*) FROM rounds;")
    final_round_count = cur.fetchone()[0]
    print(f"✅ Final verification: {final_round_count} rounds in database")
    
    # Show top standings
    cur.execute("SELECT name, points FROM standings ORDER BY points DESC LIMIT 3;")
    top3 = cur.fetchall()
    print("\n🏆 Top 3 standings:")
    for name, points in top3:
        print(f"  {name}: {points} pts")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"❌ Error: {e}")
    import traceback
    traceback.print_exc()
    sys.exit(1)
